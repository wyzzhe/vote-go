package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"vote-system/models"
	"vote-system/websocket"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移测试表
	db.AutoMigrate(&models.Poll{}, &models.Option{}, &models.Vote{})
	return db
}

func setupTestData(db *gorm.DB) (models.Poll, []models.Option) {
	// 创建测试投票问卷
	poll := models.Poll{
		Title:       "测试投票",
		Description: "这是一个测试投票",
		IsActive:    true,
	}
	db.Create(&poll)

	// 创建测试选项
	options := []models.Option{
		{PollID: poll.ID, Text: "选项1", VoteCount: 0},
		{PollID: poll.ID, Text: "选项2", VoteCount: 0},
		{PollID: poll.ID, Text: "选项3", VoteCount: 0},
	}

	for i := range options {
		db.Create(&options[i])
	}

	return poll, options
}

func TestNewPollHandler(t *testing.T) {
	db := setupTestDB()
	hub := websocket.NewHub()

	handler := NewPollHandler(db, hub)

	if handler == nil {
		t.Fatal("PollHandler不应该为nil")
	}

	if handler.db != db {
		t.Error("数据库连接设置不正确")
	}

	if handler.hub != hub {
		t.Error("WebSocket Hub设置不正确")
	}
}

func TestGetPoll_Success(t *testing.T) {
	db := setupTestDB()
	hub := websocket.NewHub()
	handler := NewPollHandler(db, hub)

	// 设置测试数据
	poll, _ := setupTestData(db)

	// 设置Gin测试模式
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/poll", handler.GetPoll)

	// 创建测试请求
	req, _ := http.NewRequest("GET", "/poll", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusOK, w.Code)
	}

	var response models.PollResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if response.Poll.ID != poll.ID {
		t.Errorf("期望投票问卷ID %d, 得到 %d", poll.ID, response.Poll.ID)
	}

	if response.Poll.Title != poll.Title {
		t.Errorf("期望标题 %s, 得到 %s", poll.Title, response.Poll.Title)
	}

	if response.TotalVotes != 0 {
		t.Errorf("期望总投票数 0, 得到 %d", response.TotalVotes)
	}

	if response.UserVoted {
		t.Error("用户不应该已投票")
	}

	if response.VotedOption != nil {
		t.Error("投票选项应该为nil")
	}
}

func TestGetPoll_NoPollFound(t *testing.T) {
	db := setupTestDB()
	hub := websocket.NewHub()
	handler := NewPollHandler(db, hub)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/poll", handler.GetPoll)

	req, _ := http.NewRequest("GET", "/poll", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusNotFound, w.Code)
	}
}

func TestVote_Success(t *testing.T) {
	db := setupTestDB()
	hub := websocket.NewHub()
	handler := NewPollHandler(db, hub)

	// 设置测试数据
	_, options := setupTestData(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/vote", handler.Vote)

	// 创建投票请求
	voteReq := models.VoteRequest{
		OptionID: options[0].ID,
	}
	jsonData, _ := json.Marshal(voteReq)

	req, _ := http.NewRequest("POST", "/vote", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusOK, w.Code)
	}

	// 验证投票记录是否创建
	var voteCount int64
	db.Model(&models.Vote{}).Count(&voteCount)
	if voteCount != 1 {
		t.Errorf("期望1条投票记录, 得到 %d", voteCount)
	}

	// 验证选项投票数是否增加
	var option models.Option
	db.First(&option, options[0].ID)
	if option.VoteCount != 1 {
		t.Errorf("期望投票数 1, 得到 %d", option.VoteCount)
	}
}

func TestVote_InvalidJSON(t *testing.T) {
	db := setupTestDB()
	hub := websocket.NewHub()
	handler := NewPollHandler(db, hub)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/vote", handler.Vote)

	req, _ := http.NewRequest("POST", "/vote", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusBadRequest, w.Code)
	}
}

func TestVote_NoPollFound(t *testing.T) {
	db := setupTestDB()
	hub := websocket.NewHub()
	handler := NewPollHandler(db, hub)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/vote", handler.Vote)

	voteReq := models.VoteRequest{
		OptionID: 999,
	}
	jsonData, _ := json.Marshal(voteReq)

	req, _ := http.NewRequest("POST", "/vote", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusNotFound, w.Code)
	}
}

func TestVote_InvalidOption(t *testing.T) {
	db := setupTestDB()
	hub := websocket.NewHub()
	handler := NewPollHandler(db, hub)

	// 设置测试数据
	setupTestData(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/vote", handler.Vote)

	voteReq := models.VoteRequest{
		OptionID: 999, // 不存在的选项ID
	}
	jsonData, _ := json.Marshal(voteReq)

	req, _ := http.NewRequest("POST", "/vote", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusBadRequest, w.Code)
	}
}

func TestVote_AlreadyVoted(t *testing.T) {
	db := setupTestDB()
	hub := websocket.NewHub()
	handler := NewPollHandler(db, hub)

	// 设置测试数据
	poll, options := setupTestData(db)

	// 先创建一个投票记录
	vote := models.Vote{
		PollID:   poll.ID,
		OptionID: options[0].ID,
		UserIP:   "127.0.0.1",
	}
	db.Create(&vote)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/vote", handler.Vote)

	voteReq := models.VoteRequest{
		OptionID: options[1].ID,
	}
	jsonData, _ := json.Marshal(voteReq)

	req, _ := http.NewRequest("POST", "/vote", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusBadRequest, w.Code)
	}
}

func TestClearVotes_Success(t *testing.T) {
	db := setupTestDB()
	hub := websocket.NewHub()
	handler := NewPollHandler(db, hub)

	// 设置测试数据
	poll, options := setupTestData(db)

	// 创建投票记录
	vote := models.Vote{
		PollID:   poll.ID,
		OptionID: options[0].ID,
		UserIP:   "127.0.0.1",
	}
	db.Create(&vote)

	// 更新选项投票数
	db.Model(&options[0]).Update("vote_count", 1)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/clear-vote", handler.ClearVotes)

	req, _ := http.NewRequest("DELETE", "/clear-vote", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusOK, w.Code)
	}

	// 验证投票记录是否删除
	var voteCount int64
	db.Model(&models.Vote{}).Count(&voteCount)
	if voteCount != 0 {
		t.Errorf("期望0条投票记录, 得到 %d", voteCount)
	}

	// 验证选项投票数是否减少
	var option models.Option
	db.First(&option, options[0].ID)
	if option.VoteCount != 0 {
		t.Errorf("期望投票数 0, 得到 %d", option.VoteCount)
	}
}

func TestClearVotes_NoVoteFound(t *testing.T) {
	db := setupTestDB()
	hub := websocket.NewHub()
	handler := NewPollHandler(db, hub)

	// 设置测试数据但不创建投票记录
	setupTestData(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/clear-vote", handler.ClearVotes)

	req, _ := http.NewRequest("DELETE", "/clear-vote", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusNotFound, w.Code)
	}
}

func TestResetPoll_Success(t *testing.T) {
	db := setupTestDB()
	hub := websocket.NewHub()
	handler := NewPollHandler(db, hub)

	// 设置测试数据
	poll, options := setupTestData(db)

	// 创建一些投票记录
	votes := []models.Vote{
		{PollID: poll.ID, OptionID: options[0].ID, UserIP: "192.168.1.1"},
		{PollID: poll.ID, OptionID: options[1].ID, UserIP: "192.168.1.2"},
		{PollID: poll.ID, OptionID: options[0].ID, UserIP: "192.168.1.3"},
	}

	for _, vote := range votes {
		db.Create(&vote)
	}

	// 更新选项投票数
	db.Model(&options[0]).Update("vote_count", 2)
	db.Model(&options[1]).Update("vote_count", 1)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/reset", handler.ResetPoll)

	req, _ := http.NewRequest("DELETE", "/reset", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusOK, w.Code)
	}

	// 验证所有投票记录是否删除
	var voteCount int64
	db.Model(&models.Vote{}).Count(&voteCount)
	if voteCount != 0 {
		t.Errorf("期望0条投票记录, 得到 %d", voteCount)
	}

	// 验证所有选项投票数是否重置为0
	var resetOptions []models.Option
	db.Where("poll_id = ?", poll.ID).Find(&resetOptions)

	for _, option := range resetOptions {
		if option.VoteCount != 0 {
			t.Errorf("选项 %s: 期望投票数 0, 得到 %d", option.Text, option.VoteCount)
		}
	}
}

func TestResetPoll_NoPollFound(t *testing.T) {
	db := setupTestDB()
	hub := websocket.NewHub()
	handler := NewPollHandler(db, hub)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/reset", handler.ResetPoll)

	req, _ := http.NewRequest("DELETE", "/reset", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusNotFound, w.Code)
	}
}
