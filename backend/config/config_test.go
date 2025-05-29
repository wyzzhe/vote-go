package config

import (
	"os"
	"testing"
)

func TestLoadConfigWithDefaults(t *testing.T) {
	// 清除环境变量以测试默认值
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")

	cfg := Load()

	// 测试默认端口
	expectedPort := "8080"
	if cfg.Port != expectedPort {
		t.Errorf("期望默认端口 %s, 得到 %s", expectedPort, cfg.Port)
	}

	// 测试默认数据库URL
	expectedDBURL := "root:password@tcp(localhost:3306)/vote_system?charset=utf8mb4&parseTime=True&loc=Local"
	if cfg.DatabaseURL != expectedDBURL {
		t.Errorf("期望默认数据库URL %s, 得到 %s", expectedDBURL, cfg.DatabaseURL)
	}
}

func TestLoadConfigWithEnvironmentVariables(t *testing.T) {
	// 设置环境变量
	testPort := "9090"
	testDBURL := "test:test@tcp(testhost:3306)/test_db"

	os.Setenv("PORT", testPort)
	os.Setenv("DATABASE_URL", testDBURL)

	// 确保测试结束后清理环境变量
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_URL")
	}()

	cfg := Load()

	// 测试从环境变量加载的端口
	if cfg.Port != testPort {
		t.Errorf("期望端口 %s, 得到 %s", testPort, cfg.Port)
	}

	// 测试从环境变量加载的数据库URL
	if cfg.DatabaseURL != testDBURL {
		t.Errorf("期望数据库URL %s, 得到 %s", testDBURL, cfg.DatabaseURL)
	}
}

func TestLoadConfigPartialEnvironmentVariables(t *testing.T) {
	// 只设置PORT环境变量，DATABASE_URL使用默认值
	testPort := "7777"
	os.Setenv("PORT", testPort)
	os.Unsetenv("DATABASE_URL")

	defer os.Unsetenv("PORT")

	cfg := Load()

	// 测试从环境变量加载的端口
	if cfg.Port != testPort {
		t.Errorf("期望端口 %s, 得到 %s", testPort, cfg.Port)
	}

	// 测试默认数据库URL
	expectedDBURL := "root:password@tcp(localhost:3306)/vote_system?charset=utf8mb4&parseTime=True&loc=Local"
	if cfg.DatabaseURL != expectedDBURL {
		t.Errorf("期望默认数据库URL %s, 得到 %s", expectedDBURL, cfg.DatabaseURL)
	}
}

func TestConfigStruct(t *testing.T) {
	// 测试Config结构体
	cfg := &Config{
		Port:        "8080",
		DatabaseURL: "test_url",
	}

	if cfg.Port != "8080" {
		t.Errorf("期望端口 8080, 得到 %s", cfg.Port)
	}

	if cfg.DatabaseURL != "test_url" {
		t.Errorf("期望数据库URL test_url, 得到 %s", cfg.DatabaseURL)
	}
}

func TestLoadConfigEmptyEnvironmentVariables(t *testing.T) {
	// 设置空的环境变量，应该使用默认值
	os.Setenv("PORT", "")
	os.Setenv("DATABASE_URL", "")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_URL")
	}()

	cfg := Load()

	// 空字符串应该被视为未设置，使用默认值
	expectedPort := "8080"
	if cfg.Port != expectedPort {
		t.Errorf("期望默认端口 %s, 得到 %s", expectedPort, cfg.Port)
	}

	expectedDBURL := "root:password@tcp(localhost:3306)/vote_system?charset=utf8mb4&parseTime=True&loc=Local"
	if cfg.DatabaseURL != expectedDBURL {
		t.Errorf("期望默认数据库URL %s, 得到 %s", expectedDBURL, cfg.DatabaseURL)
	}
}
