package queue

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"

	"notiair/internal/config"
	"notiair/internal/routing"
	"notiair/internal/transport/http/telegram"
)

type Worker struct {
	server *asynq.Server
	bot    *telegram.Client
	cfg    config.QueueConfig
}

type WorkerOptions struct {
	Concurrency int
}

func NewWorker(cfg config.QueueConfig, bot *telegram.Client, opts WorkerOptions) *Worker {
	server := asynq.NewServer(asynq.RedisClientOpt{Addr: cfg.URL}, asynq.Config{
		Concurrency: opts.Concurrency,
		Queues: map[string]int{
			cfg.Namespace: 1,
		},
	})

	return &Worker{server: server, bot: bot, cfg: cfg}
}

func (w *Worker) Start() error {
	handler := func(ctx context.Context, task *asynq.Task) error {
		var payload routing.Task
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return err
		}

		return w.bot.SendMessage(ctx, payload.ChannelID, payload)
	}

	mux := asynq.NewServeMux()
	mux.HandleFunc("notification:deliver", handler)

	log.Printf("queue worker started (namespace=%s)", w.cfg.Namespace)
	return w.server.Start(mux)
}

func (w *Worker) Shutdown() {
	w.server.Shutdown()
}
