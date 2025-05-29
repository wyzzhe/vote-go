package models

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移测试表
	db.AutoMigrate(&Poll{}, &Option{}, &Vote{})
	return db
}

func TestPollModel(t *testing.T) {
	db := setupTestDB()

	// 测试创建投票问卷
	poll := Poll{
		Title:       "测试投票",
		Description: "这是一个测试投票",
		IsActive:    true,
	}

	result := db.Create(&poll)
	if result.Error != nil {
		t.Fatalf("创建投票问卷失败: %v", result.Error)
	}

	if poll.ID == 0 {
		t.Error("投票问卷ID应该被自动设置")
	}

	if poll.CreatedAt.IsZero() {
		t.Error("CreatedAt应该被自动设置")
	}

	// 测试查询投票问卷
	var foundPoll Poll
	db.First(&foundPoll, poll.ID)

	if foundPoll.Title != poll.Title {
		t.Errorf("期望标题 %s, 得到 %s", poll.Title, foundPoll.Title)
	}

	if foundPoll.Description != poll.Description {
		t.Errorf("期望描述 %s, 得到 %s", poll.Description, foundPoll.Description)
	}

	if foundPoll.IsActive != poll.IsActive {
		t.Errorf("期望活跃状态 %v, 得到 %v", poll.IsActive, foundPoll.IsActive)
	}
}

func TestOptionModel(t *testing.T) {
	db := setupTestDB()

	// 先创建一个投票问卷
	poll := Poll{
		Title:    "测试投票",
		IsActive: true,
	}
	db.Create(&poll)

	// 测试创建选项
	option := Option{
		PollID:    poll.ID,
		Text:      "选项1",
		VoteCount: 0,
	}

	result := db.Create(&option)
	if result.Error != nil {
		t.Fatalf("创建选项失败: %v", result.Error)
	}

	if option.ID == 0 {
		t.Error("选项ID应该被自动设置")
	}

	// 测试查询选项
	var foundOption Option
	db.First(&foundOption, option.ID)

	if foundOption.Text != option.Text {
		t.Errorf("期望选项文本 %s, 得到 %s", option.Text, foundOption.Text)
	}

	if foundOption.PollID != poll.ID {
		t.Errorf("期望投票问卷ID %d, 得到 %d", poll.ID, foundOption.PollID)
	}

	if foundOption.VoteCount != 0 {
		t.Errorf("期望投票数 0, 得到 %d", foundOption.VoteCount)
	}
}

func TestVoteModel(t *testing.T) {
	db := setupTestDB()

	// 创建测试数据
	poll := Poll{Title: "测试投票", IsActive: true}
	db.Create(&poll)

	option := Option{PollID: poll.ID, Text: "选项1", VoteCount: 0}
	db.Create(&option)

	// 测试创建投票记录
	vote := Vote{
		PollID:   poll.ID,
		OptionID: option.ID,
		UserIP:   "192.168.1.1",
	}

	result := db.Create(&vote)
	if result.Error != nil {
		t.Fatalf("创建投票记录失败: %v", result.Error)
	}

	if vote.ID == 0 {
		t.Error("投票记录ID应该被自动设置")
	}

	// 测试查询投票记录
	var foundVote Vote
	db.First(&foundVote, vote.ID)

	if foundVote.PollID != poll.ID {
		t.Errorf("期望投票问卷ID %d, 得到 %d", poll.ID, foundVote.PollID)
	}

	if foundVote.OptionID != option.ID {
		t.Errorf("期望选项ID %d, 得到 %d", option.ID, foundVote.OptionID)
	}

	if foundVote.UserIP != vote.UserIP {
		t.Errorf("期望用户IP %s, 得到 %s", vote.UserIP, foundVote.UserIP)
	}
}

func TestPollWithOptions(t *testing.T) {
	db := setupTestDB()

	// 创建投票问卷
	poll := Poll{
		Title:    "测试投票",
		IsActive: true,
	}
	db.Create(&poll)

	// 创建多个选项
	options := []Option{
		{PollID: poll.ID, Text: "选项1", VoteCount: 0},
		{PollID: poll.ID, Text: "选项2", VoteCount: 5},
		{PollID: poll.ID, Text: "选项3", VoteCount: 3},
	}

	for _, option := range options {
		db.Create(&option)
	}

	// 测试预加载选项
	var foundPoll Poll
	db.Preload("Options").First(&foundPoll, poll.ID)

	if len(foundPoll.Options) != 3 {
		t.Errorf("期望3个选项, 得到 %d", len(foundPoll.Options))
	}

	// 验证选项内容
	expectedTexts := []string{"选项1", "选项2", "选项3"}
	expectedCounts := []int{0, 5, 3}

	for i, option := range foundPoll.Options {
		if option.Text != expectedTexts[i] {
			t.Errorf("选项 %d: 期望文本 %s, 得到 %s", i, expectedTexts[i], option.Text)
		}
		if option.VoteCount != expectedCounts[i] {
			t.Errorf("选项 %d: 期望投票数 %d, 得到 %d", i, expectedCounts[i], option.VoteCount)
		}
	}
}

func TestVoteRequest(t *testing.T) {
	// 测试投票请求结构
	req := VoteRequest{
		OptionID: 123,
	}

	if req.OptionID != 123 {
		t.Errorf("期望选项ID 123, 得到 %d", req.OptionID)
	}
}

func TestPollResponse(t *testing.T) {
	// 测试投票响应结构
	poll := Poll{
		ID:       1,
		Title:    "测试投票",
		IsActive: true,
	}

	votedOption := uint(2)
	response := PollResponse{
		Poll:        poll,
		TotalVotes:  10,
		UserVoted:   true,
		VotedOption: &votedOption,
	}

	if response.Poll.ID != 1 {
		t.Errorf("期望投票问卷ID 1, 得到 %d", response.Poll.ID)
	}

	if response.TotalVotes != 10 {
		t.Errorf("期望总投票数 10, 得到 %d", response.TotalVotes)
	}

	if !response.UserVoted {
		t.Error("期望用户已投票")
	}

	if response.VotedOption == nil || *response.VotedOption != 2 {
		t.Errorf("期望投票选项ID 2, 得到 %v", response.VotedOption)
	}
}
