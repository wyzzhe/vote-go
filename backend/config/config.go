package config

import (
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "root:password@tcp(localhost:3306)/vote_system?charset=utf8mb4&parseTime=True&loc=Local"
	}

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
	}
}
