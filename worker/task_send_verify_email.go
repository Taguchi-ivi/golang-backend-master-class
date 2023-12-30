package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"input-backend-master-class/util"

	"github.com/hibiken/asynq"
)

// タスクの種類を定義
const TaskSendVerifyEmail = "task:send_verify_email"

type PayloadSendVerifyEmail struct {
	UserName string `json:"user_name"`
}

func (distribute *RedisTaskDistributor) DistributeTaskSendVerifyEmail(ctx context.Context, payload *PayloadSendVerifyEmail, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)
	info, err := distribute.client.Enqueue(task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	fmt.Println("enqueued task", info.ID)
	// log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
	// 	Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	user, err := processor.store.GetUser(ctx, payload.UserName)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user doesn't exist: %w", asynq.SkipRetry)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	verifyEmail, err := processor.store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
		UserName:   user.UserName,
		Email:      user.Email,
		SecretCode: util.RandomString(32),
	})
	if err != nil {
		return fmt.Errorf("failed to create verify email: %w", err)
	}

	subject := "welcome to web app"
	verifyUrl := fmt.Sprintf("http://localhost:8080/v1/verify_email?email_id=%s&secret_code=%s", &verifyEmail.ID, verifyEmail.SecretCode)
	content := fmt.Sprintf(`
		Hello %s, </br>
		Thank you for registering with us! </br>
		Please <a href="%s">click here</a> to verify your email address. </br>
		`, user.FullName, verifyUrl)
	to := []string{user.Email}

	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	// email := mail.NewGmailSender(user.UserName, )
	return nil
}
