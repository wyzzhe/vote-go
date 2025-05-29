package database

import (
	"testing"
	"vote-system/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInit(t *testing.T) {
	// 使用内存SQLite数据库进行测试
	db, err := Init("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("数据库初始化失败: %v", err)
	}

	if db == nil {
		t.Fatal("数据库连接不应该为nil")
	}

	// 验证表是否已创建
	if !db.Migrator().HasTable(&models.Poll{}) {
		t.Error("Poll表应该已创建")
	}

	if !db.Migrator().HasTable(&models.Option{}) {
		t.Error("Option表应该已创建")
	}

	if !db.Migrator().HasTable(&models.Vote{}) {
		t.Error("Vote表应该已创建")
	}
}

func TestInitDefaultData(t *testing.T) {
	// 创建测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}

	// 迁移表结构
	db.AutoMigrate(&models.Poll{}, &models.Option{}, &models.Vote{})

	// 调用初始化默认数据函数
	initDefaultData(db)

	// 验证是否创建了默认投票问卷
	var pollCount int64
	db.Model(&models.Poll{}).Count(&pollCount)
	if pollCount != 1 {
		t.Errorf("期望1个投票问卷, 得到 %d", pollCount)
	}

	// 验证投票问卷内容
	var poll models.Poll
	db.First(&poll)

	expectedTitle := "您最喜欢的编程语言是什么？"
	if poll.Title != expectedTitle {
		t.Errorf("期望标题 %s, 得到 %s", expectedTitle, poll.Title)
	}

	expectedDescription := "请选择您最喜欢的编程语言"
	if poll.Description != expectedDescription {
		t.Errorf("期望描述 %s, 得到 %s", expectedDescription, poll.Description)
	}

	if !poll.IsActive {
		t.Error("投票问卷应该是活跃状态")
	}

	// 验证是否创建了选项
	var optionCount int64
	db.Model(&models.Option{}).Count(&optionCount)
	if optionCount != 5 {
		t.Errorf("期望5个选项, 得到 %d", optionCount)
	}

	// 验证选项内容
	var options []models.Option
	db.Where("poll_id = ?", poll.ID).Find(&options)

	expectedOptions := []string{"Go", "Python", "JavaScript", "Java", "TypeScript"}
	if len(options) != len(expectedOptions) {
		t.Errorf("期望 %d 个选项, 得到 %d", len(expectedOptions), len(options))
	}

	for i, option := range options {
		if i < len(expectedOptions) && option.Text != expectedOptions[i] {
			t.Errorf("选项 %d: 期望 %s, 得到 %s", i, expectedOptions[i], option.Text)
		}

		if option.VoteCount != 0 {
			t.Errorf("选项 %s: 期望投票数 0, 得到 %d", option.Text, option.VoteCount)
		}

		if option.PollID != poll.ID {
			t.Errorf("选项 %s: 期望投票问卷ID %d, 得到 %d", option.Text, poll.ID, option.PollID)
		}
	}
}

func TestInitDefaultDataWithExistingPoll(t *testing.T) {
	// 创建测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}

	// 迁移表结构
	db.AutoMigrate(&models.Poll{}, &models.Option{}, &models.Vote{})

	// 先创建一个投票问卷
	existingPoll := models.Poll{
		Title:    "已存在的投票",
		IsActive: true,
	}
	db.Create(&existingPoll)

	// 调用初始化默认数据函数
	initDefaultData(db)

	// 验证没有创建新的投票问卷
	var pollCount int64
	db.Model(&models.Poll{}).Count(&pollCount)
	if pollCount != 1 {
		t.Errorf("期望1个投票问卷, 得到 %d", pollCount)
	}

	// 验证投票问卷仍然是原来的
	var poll models.Poll
	db.First(&poll)
	if poll.Title != "已存在的投票" {
		t.Errorf("期望标题 '已存在的投票', 得到 %s", poll.Title)
	}
}

func TestInitWithInvalidDatabaseURL(t *testing.T) {
	// 测试无效的数据库URL
	_, err := Init("invalid://database/url")
	if err == nil {
		t.Error("期望数据库初始化失败，但没有返回错误")
	}
}

func TestDatabaseMigration(t *testing.T) {
	// 创建测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}

	// 手动执行迁移
	err = db.AutoMigrate(
		&models.Poll{},
		&models.Option{},
		&models.Vote{},
	)
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}

	// 验证表结构
	if !db.Migrator().HasTable("polls") {
		t.Error("polls表应该已创建")
	}

	if !db.Migrator().HasTable("options") {
		t.Error("options表应该已创建")
	}

	if !db.Migrator().HasTable("votes") {
		t.Error("votes表应该已创建")
	}

	// 验证列是否存在
	if !db.Migrator().HasColumn(&models.Poll{}, "title") {
		t.Error("polls表应该有title列")
	}

	if !db.Migrator().HasColumn(&models.Option{}, "poll_id") {
		t.Error("options表应该有poll_id列")
	}

	if !db.Migrator().HasColumn(&models.Vote{}, "user_ip") {
		t.Error("votes表应该有user_ip列")
	}
}
