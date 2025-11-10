package queue

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"

	"notiair/internal/config"
	"notiair/internal/routing"
)

type Client interface {
	Enqueue(ctx context.Context, task routing.Task) error
	Close() error
}

type asynqClient struct {
	client *asynq.Client
	cfg    config.QueueConfig
}

func NewAsynqClient(cfg config.QueueConfig) Client {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.URL})
	return &asynqClient{client: client, cfg: cfg}
}

func (c *asynqClient) Enqueue(ctx context.Context, task routing.Task) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return err
	}

	opts := []asynq.Option{
		asynq.MaxRetry(c.cfg.RetryLimit),
		asynq.Queue(c.cfg.Namespace),
	}

	job := asynq.NewTask("notification:deliver", payload)
	_, err = c.client.EnqueueContext(ctx, job, opts...)
	return err
}

func (c *asynqClient) Close() error {
	return c.client.Close()
}
