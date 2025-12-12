//go:build asynq_example

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hibiken/asynq"
)

const (
	taskTypeWelcome  = "email:welcome"
	redisAddrDefault = "127.0.0.1:6379"
)

type welcomePayload struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}

func main() {
	mode := flag.String("mode", "enqueue", "mode: enqueue or worker")
	redisAddr := flag.String("redis", envOr("REDIS_ADDR", redisAddrDefault), "redis address host:port")
	flag.Parse()

	switch *mode {
	case "enqueue":
		runEnqueue(*redisAddr)
	case "worker":
		runWorker(*redisAddr)
	default:
		log.Fatalf("unknown mode %q (use enqueue|worker)", *mode)
	}
}

// 生产者
func runEnqueue(redisAddr string) {
	sched := asynq.NewScheduler(asynq.RedisClientOpt{Addr: redisAddr}, &asynq.SchedulerOpts{})

	payload := welcomePayload{UserID: 123, Email: "user@example.com"}
	data, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("marshal payload: %v", err)
	}

	task := asynq.NewTask(taskTypeWelcome, data)
	entryID, err := sched.Register("*/1 * * * *", task,
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(10*time.Second),
	)
	if err != nil {
		log.Fatalf("register schedule: %v", err)
	}

	log.Printf("scheduled task every minute entry=%s queue=default", entryID)
	if err := sched.Run(); err != nil {
		log.Fatalf("run scheduler: %v", err)
	}
}

// 消费者
func runWorker(redisAddr string) {
	srv := asynq.NewServer(asynq.RedisClientOpt{Addr: redisAddr}, asynq.Config{
		Concurrency: 1,
	})

	mux := asynq.NewServeMux()
	mux.HandleFunc(taskTypeWelcome, handleWelcomeEmail)

	if err := srv.Run(mux); err != nil && !errors.Is(err, asynq.ErrServerClosed) {
		log.Fatalf("run server: %v", err)
	}
}

func handleWelcomeEmail(ctx context.Context, task *asynq.Task) error {
	var payload welcomePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	log.Printf("[worker] send welcome email to user=%d email=%s", payload.UserID, payload.Email)
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
