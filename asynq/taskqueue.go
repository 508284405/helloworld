package asynq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TaskTypeWelcome = "email:welcome"
	DefaultQueue    = "default"
	DefaultCron     = "*/1 * * * *"
)

type WelcomePayload struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}

// TaskQueue wraps Asynq server and scheduler lifecycle.
type TaskQueue struct {
	redisOpt  asynq.RedisClientOpt
	server    *asynq.Server
	scheduler *asynq.Scheduler
	mux       *asynq.ServeMux
	onceStart sync.Once
	onceStop  sync.Once
}

func New(redisAddr string) *TaskQueue {
	opt := asynq.RedisClientOpt{Addr: redisAddr}
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskTypeWelcome, handleWelcomeEmail)

	return &TaskQueue{
		redisOpt:  opt,
		server:    asynq.NewServer(opt, asynq.Config{Concurrency: 1}),
		scheduler: asynq.NewScheduler(opt, &asynq.SchedulerOpts{}),
		mux:       mux,
	}
}

func (tq *TaskQueue) Start() {
	tq.onceStart.Do(func() {
		go func() {
			if err := tq.server.Run(tq.mux); err != nil && err != asynq.ErrServerClosed {
				log.Fatalf("asynq server: %v", err)
			}
		}()

		go func() {
			if err := tq.scheduler.Run(); err != nil {
				log.Fatalf("asynq scheduler: %v", err)
			}
		}()
	})
}

func (tq *TaskQueue) Shutdown(ctx context.Context) {
	tq.onceStop.Do(func() {
		done := make(chan struct{})
		go func() {
			tq.scheduler.Shutdown()
			tq.server.Shutdown()
			close(done)
		}()

		select {
		case <-done:
		case <-ctx.Done():
			log.Printf("asynq shutdown timeout: %v", ctx.Err())
		}
	})
}

func (tq *TaskQueue) RegisterWelcomeSchedule(ctx context.Context, payload WelcomePayload, cronExpr string) (string, error) {
	if payload.UserID == 0 || payload.Email == "" {
		return "", fmt.Errorf("user_id and email are required")
	}

	if cronExpr == "" {
		cronExpr = DefaultCron
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode payload: %w", err)
	}

	task := asynq.NewTask(TaskTypeWelcome, data)
	entryID, err := tq.scheduler.Register(cronExpr, task,
		asynq.Queue(DefaultQueue),
		asynq.MaxRetry(3),
		asynq.Timeout(10*time.Second),
	)
	if err != nil {
		return "", fmt.Errorf("register schedule: %w", err)
	}

	return entryID, nil
}

func handleWelcomeEmail(ctx context.Context, task *asynq.Task) error {
	var payload WelcomePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	log.Printf("[worker] send welcome email to user=%d email=%s", payload.UserID, payload.Email)
	return nil
}
