package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	appasynq "helloworld/asynq"
)

const redisAddrDefault = "127.0.0.1:6379"

type scheduleRequest struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Cron   string `json:"cron"`
}

func main() {
	redisAddr := envOr("REDIS_ADDR", redisAddrDefault)

	queue := appasynq.New(redisAddr)
	queue.Start()
	defer queue.Shutdown(context.Background())

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello world")
	})

	r.POST("/tasks/welcome", func(c *gin.Context) {
		var req scheduleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}
		if req.UserID == 0 || req.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and email are required"})
			return
		}

		cronExpr := req.Cron
		if cronExpr == "" {
			cronExpr = appasynq.DefaultCron
		}

		payload := appasynq.WelcomePayload{UserID: req.UserID, Email: req.Email}
		entryID, err := queue.RegisterWelcomeSchedule(c.Request.Context(), payload, cronExpr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"entry_id": entryID,
			"cron":     cronExpr,
			"type":     appasynq.TaskTypeWelcome,
		})
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("gin server: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
