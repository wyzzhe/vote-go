package database

import (
	"vote-system/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Init(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 自动迁移数据库表
	err = db.AutoMigrate(
		&models.Poll{},
		&models.Option{},
		&models.Vote{},
	)
	if err != nil {
		return nil, err
	}

	// 初始化默认数据
	initDefaultData(db)

	return db, nil
}

func initDefaultData(db *gorm.DB) {
	// 检查是否已有投票问卷
	var count int64
	db.Model(&models.Poll{}).Count(&count)
	if count > 0 {
		return
	}

	// 创建默认投票问卷
	poll := models.Poll{
		Title:       "您最喜欢的编程语言是什么？",
		Description: "请选择您最喜欢的编程语言",
		IsActive:    true,
	}

	result := db.Create(&poll)
	if result.Error != nil {
		return
	}

	// 创建选项
	options := []models.Option{
		{PollID: poll.ID, Text: "Go", VoteCount: 0},
		{PollID: poll.ID, Text: "Python", VoteCount: 0},
		{PollID: poll.ID, Text: "JavaScript", VoteCount: 0},
		{PollID: poll.ID, Text: "Java", VoteCount: 0},
		{PollID: poll.ID, Text: "TypeScript", VoteCount: 0},
	}

	for _, option := range options {
		db.Create(&option)
	}
}
