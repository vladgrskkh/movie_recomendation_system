package main

import (
	"fmt"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/vladgrskkh/movie_recomendation_system/notificationservice/internal/mailer"
)

// SendEmailHandler struct implements consumer.Handler interface.
// This struct is used with consumers that reads messages, logs them and sends email to recepient.
type SendEmailHandler struct {
	mailer *mailer.Mailer
}

func (app *application) NewSendEmailHandler() *SendEmailHandler {
	return &SendEmailHandler{
		mailer: app.mailer,
	}
}

// HandlerMessage func send email to repepient with either activation token or reset password token.
func (h *SendEmailHandler) HandleMessage(message []byte, topic kafka.TopicPartition, consumerNumber int) error {
	var details struct {
		UserID       int64   `json:"user_id"`
		Email        string  `json:"email"`
		Name         string  `json:"name"`
		Token        string  `json:"token"`
		TemplateName *string `json:"template_name,omitempty"`
		Task         string  `json:"task"`
	}

	err := readJSON(message, &details)
	if err != nil {
		return fmt.Errorf("error unmarshaling json: %s", err.Error())
	}

	msg := fmt.Sprintf("Consumer %d, Message from kafka with offset %d task:'%s' on partition %d", consumerNumber, topic.Offset, details.Task, topic.Partition)
	slog.Info(msg)

	if details.TemplateName == nil {
		return nil
	}

	err = h.mailer.Send(details.Email, *details.TemplateName, map[string]string{
		"Token": details.Token,
		"name":  details.Name,
	})
	if err != nil {
		return fmt.Errorf("error sending mail: %s", err.Error())
	}
	return nil
}
