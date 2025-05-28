package models

import (
	"time"

	"gorm.io/gorm"
)

// Poll 投票问卷模型
type Poll struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Title       string         `gorm:"size:255;not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	Options     []Option       `gorm:"foreignKey:PollID" json:"options"`
}

// Option 选项模型
type Option struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	PollID    uint           `gorm:"not null" json:"poll_id"`
	Text      string         `gorm:"size:255;not null" json:"text"`
	VoteCount int            `gorm:"default:0" json:"vote_count"`
}

// Vote 投票记录模型
type Vote struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	PollID    uint           `gorm:"not null" json:"poll_id"`
	OptionID  uint           `gorm:"not null" json:"option_id"`
	UserIP    string         `gorm:"size:45" json:"user_ip"`
}

// VoteRequest 投票请求结构
type VoteRequest struct {
	OptionID uint `json:"option_id" binding:"required"`
}

// PollResponse 投票问卷响应结构
type PollResponse struct {
	Poll        Poll  `json:"poll"`
	TotalVotes  int   `json:"total_votes"`
	UserVoted   bool  `json:"user_voted"`
	VotedOption *uint `json:"voted_option,omitempty"`
}
