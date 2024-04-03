package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	db "github.com/thaovo29/simplebank/db/sqlc"
	"github.com/thaovo29/simplebank/util"
)

const TaskSendVerifynEmail = "task:send_verify_email"

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSEndVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("fail to marshal task payload: %w", err)
	}
	task := asynq.NewTask(TaskSendVerifynEmail, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("fail to enqueue task: %w", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_entry", info.MaxRetry).Msg("enqueued task")
	return nil
}

func (processor *RedisTaskProccessor) ProcessTaskSendVerifynEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("fail to unmarshal payload: %w", asynq.SkipRetry)
	}

	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		// if  errors.Is(err, db.ErrRecordNotFound) {
		// 	return fmt.Errorf("user doesn't exist: %w", asynq.SkipRetry)
		// }
		return fmt.Errorf("fail to get user: %w", err)
	}

	verifyEmail, err := processor.store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
		Username:   user.Username,
		Email:      user.Email,
		SecretCode: util.RandomString(32),
	})

	if err != nil {
		return fmt.Errorf("fail to create verify email: %w", err)
	}

	subject := "Welcome to Simple Bank"
	verifyURL := fmt.Sprintf("http://localhost:8080/v1/verify_email?email_id=%d&secret_code=%s", verifyEmail.ID, verifyEmail.SecretCode)
	content := fmt.Sprintf(`Hello: %s <br/>
		Thank you for registering with us! <br/>
		Please <a href="%s">Click here</a> to vrify email <br/>
	`, user.FullName, verifyURL)

	to := []string{user.Email}
	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("fail to send verify email")
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).
		Msg("Processed Task")
	return nil
}
