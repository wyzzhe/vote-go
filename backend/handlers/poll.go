package handlers

import (
	"net/http"
	"vote-system/models"
	"vote-system/websocket"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PollHandler struct {
	db  *gorm.DB
	hub *websocket.Hub
}

func NewPollHandler(db *gorm.DB, hub *websocket.Hub) *PollHandler {
	return &PollHandler{
		db:  db,
		hub: hub,
	}
}

// GetPoll 获取投票问卷和统计数据
func (h *PollHandler) GetPoll(c *gin.Context) {
	userIP := c.ClientIP()

	// 获取活跃的投票问卷
	var poll models.Poll
	if err := h.db.Where("is_active = ?", true).Preload("Options").First(&poll).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active poll found"})
		return
	}

	// 计算总票数
	totalVotes := 0
	for _, option := range poll.Options {
		totalVotes += option.VoteCount
	}

	// 检查用户是否已投票
	var vote models.Vote
	userVoted := false
	var votedOption *uint

	if err := h.db.Where("poll_id = ? AND user_ip = ?", poll.ID, userIP).First(&vote).Error; err == nil {
		userVoted = true
		votedOption = &vote.OptionID
	}

	response := models.PollResponse{
		Poll:        poll,
		TotalVotes:  totalVotes,
		UserVoted:   userVoted,
		VotedOption: votedOption,
	}

	c.JSON(http.StatusOK, response)
}

// Vote 提交投票
func (h *PollHandler) Vote(c *gin.Context) {
	userIP := c.ClientIP()

	var req models.VoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取活跃的投票问卷
	var poll models.Poll
	if err := h.db.Where("is_active = ?", true).First(&poll).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active poll found"})
		return
	}

	// 检查选项是否存在
	var option models.Option
	if err := h.db.Where("id = ? AND poll_id = ?", req.OptionID, poll.ID).First(&option).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid option"})
		return
	}

	// 检查用户是否已投票
	var existingVote models.Vote
	if err := h.db.Where("poll_id = ? AND user_ip = ?", poll.ID, userIP).First(&existingVote).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already voted"})
		return
	}

	// 开始事务
	tx := h.db.Begin()

	// 创建投票记录
	vote := models.Vote{
		PollID:   poll.ID,
		OptionID: req.OptionID,
		UserIP:   userIP,
	}

	if err := tx.Create(&vote).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create vote"})
		return
	}

	// 增加选项投票数
	if err := tx.Model(&option).Update("vote_count", gorm.Expr("vote_count + ?", 1)).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vote count"})
		return
	}

	// 提交事务
	tx.Commit()

	// 获取更新后的数据
	h.db.Where("is_active = ?", true).Preload("Options").First(&poll)

	// 广播更新
	h.hub.BroadcastPollUpdate(poll)

	c.JSON(http.StatusOK, gin.H{"message": "Vote submitted successfully"})
}

// ClearVotes 清除当前用户的投票记录（仅开发模式）
func (h *PollHandler) ClearVotes(c *gin.Context) {
	userIP := c.ClientIP()

	// 获取活跃的投票问卷
	var poll models.Poll
	if err := h.db.Where("is_active = ?", true).First(&poll).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active poll found"})
		return
	}

	// 查找用户的投票记录
	var vote models.Vote
	if err := h.db.Where("poll_id = ? AND user_ip = ?", poll.ID, userIP).First(&vote).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No vote found for this user"})
		return
	}

	// 开始事务
	tx := h.db.Begin()

	// 减少选项的投票数
	var option models.Option
	if err := tx.Where("id = ?", vote.OptionID).First(&option).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find option"})
		return
	}

	if err := tx.Model(&option).Update("vote_count", gorm.Expr("vote_count - ?", 1)).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vote count"})
		return
	}

	// 删除投票记录
	if err := tx.Delete(&vote).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vote"})
		return
	}

	// 提交事务
	tx.Commit()

	// 获取更新后的数据并广播
	h.db.Where("is_active = ?", true).Preload("Options").First(&poll)
	h.hub.BroadcastPollUpdate(poll)

	c.JSON(http.StatusOK, gin.H{"message": "Vote cleared successfully"})
}

// ResetPoll 重置投票（清除所有投票记录）
func (h *PollHandler) ResetPoll(c *gin.Context) {
	// 获取活跃的投票问卷
	var poll models.Poll
	if err := h.db.Where("is_active = ?", true).Preload("Options").First(&poll).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active poll found"})
		return
	}

	// 开始事务
	tx := h.db.Begin()

	// 删除所有投票记录
	if err := tx.Where("poll_id = ?", poll.ID).Delete(&models.Vote{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear votes"})
		return
	}

	// 重置所有选项的投票数
	if err := tx.Model(&models.Option{}).Where("poll_id = ?", poll.ID).Update("vote_count", 0).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset vote counts"})
		return
	}

	// 提交事务
	tx.Commit()

	// 获取更新后的数据并广播
	h.db.Where("is_active = ?", true).Preload("Options").First(&poll)
	h.hub.BroadcastPollUpdate(poll)

	c.JSON(http.StatusOK, gin.H{"message": "Poll reset successfully"})
}
