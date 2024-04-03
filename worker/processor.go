package worker

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	db "github.com/thaovo29/simplebank/db/sqlc"
	"github.com/thaovo29/simplebank/mail"
)

const (
	QueueCritcal = "critical"
	QueueDefault = "default"
)

type TaskProcessor interface {
	Start() error
	Shutdown()
	ProcessTaskSendVerifynEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProccessor struct {
	server *asynq.Server
	store  db.Store
	mailer mail.EmailSender
}

func NewTaskProccessor(redisOpt asynq.RedisClientOpt, store db.Store, mailer mail.EmailSender) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritcal: 10,
				QueueDefault: 5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().Err(err).Str("byte", task.Type()).
					Bytes("payload", task.Payload()).Msg("process task failed")
			}),
			Logger: NewLogger(),
		},
	)
	return &RedisTaskProccessor{
		server,
		store,
		mailer,
	}
}

func (processor *RedisTaskProccessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskSendVerifynEmail, processor.ProcessTaskSendVerifynEmail)
	return processor.server.Start(mux)
}

func (processor *RedisTaskProccessor) Shutdown() {
	processor.server.Shutdown()
}