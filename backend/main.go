package main

import (
	"log"
	"net/http"
	"vote-system/config"
	"vote-system/database"
	"vote-system/handlers"
	"vote-system/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	cfg := config.Load()

	// 初始化数据库
	db, err := database.Init(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 初始化WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// 设置Gin路由
	r := gin.Default()

	// CORS配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Session-ID"},
		AllowCredentials: true,
	}))

	// 创建handlers
	pollHandler := handlers.NewPollHandler(db, hub)

	// API路由
	api := r.Group("/api")
	{
		api.GET("/poll", pollHandler.GetPoll)
		api.POST("/poll/vote", pollHandler.Vote)
		api.DELETE("/poll/clear-my-vote", pollHandler.ClearVotes)
		api.DELETE("/poll/reset", pollHandler.ResetPoll)
	}

	// WebSocket路由
	r.GET("/ws/poll", func(c *gin.Context) {
		websocket.ServeWS(hub, c.Writer, c.Request)
	})

	// 启动服务器
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
