package consumer

import (
	"log/slog"
	"os"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	sessionTimeout = 7000
	noTimeout      = -1
)

type Handler interface {
	HandleMessage(message []byte, topic kafka.TopicPartition, consumerNumber int) error
}

type Consumer struct {
	consumer       *kafka.Consumer
	handler        Handler
	logger         *slog.Logger
	stop           bool
	consumerNumber int
}

func NewConsumer(logger *slog.Logger, handler Handler, topic, consumerGroup string, consumerNumber int) (*Consumer, error) {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers":        os.Getenv("KAFKA_ADDRESS"),
		"group.id":                 consumerGroup,
		"session.timeout.ms":       sessionTimeout,
		"enable.auto.offset.store": false,
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  5000,
		"auto.offset.reset":        "earliest", // might need to change to latest

		// SASL_SSL
		"security.protocol":        "SASL_SSL",
		"ssl.ca.location":          "./cert/ca-root.pem",
		"ssl.certificate.location": "./cert/client-certificate.pem",
		"ssl.key.location":         "./cert/client-private-key.pem",
		"ssl.key.password":         os.Getenv("KAFKA_PASSWORD_SSL"),
		"sasl.mechanisms":          "PLAIN",
		"sasl.username":            os.Getenv("KAFKA_USERNAME"),
		"sasl.password":            os.Getenv("KAFKA_PASSWORD_USER"),
	}

	c, err := kafka.NewConsumer(cfg)
	if err != nil {
		return nil, err
	}

	err = c.Subscribe(topic, nil)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer:       c,
		handler:        handler,
		logger:         logger,
		consumerNumber: consumerNumber,
	}, nil
}

func (c *Consumer) Start() {
	for !c.stop {
		c.logger.Info("Reading message from kafka")
		kafkaMessage, err := c.consumer.ReadMessage(noTimeout)
		if err != nil {
			c.logger.Error(err.Error())
		}

		if kafkaMessage == nil {
			c.logger.Info("Message is nil")
			continue
		}

		c.logger.Info("Message is not nil")

		// NOTE: need to experiment with this(when i failed to send email what do i do(mb dlq or just ignore it and let
		// user handler this by calling retry send email himself))
		// for now i will do second cause its easier
		err = c.handler.HandleMessage(kafkaMessage.Value, kafkaMessage.TopicPartition, c.consumerNumber)
		if err != nil {
			// NOTE: think about dlq
			c.logger.Error(err.Error())
		}

		c.logger.Info("Message handled")

		_, err = c.consumer.StoreMessage(kafkaMessage)
		if err != nil {
			c.logger.Error(err.Error())
			continue
		}
	}
}

func (c *Consumer) Stop() error {
	c.stop = true

	_, err := c.consumer.Commit()
	if err != nil {
		return err
	}

	c.logger.Info("Commited offset")
	return c.consumer.Close()
}
