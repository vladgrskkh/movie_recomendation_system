package consumer

import (
	"log/slog"
	"strings"

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
	stop           bool
	consumerNumber int
}

func NewConsumer(handler Handler, address []string, topic, consumerGroup string, consumerNumber int) (*Consumer, error) {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers":        strings.Join(address, ","),
		"group.id":                 consumerGroup,
		"session.timeout.ms":       sessionTimeout,
		"enable.auto.offset.store": false,
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  5000,
		"auto.offset.reset":        "earliest", // might need to change to latest
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
		consumerNumber: consumerNumber,
	}, nil
}

func (c *Consumer) Start() {
	for !c.stop {
		kafkaMessage, err := c.consumer.ReadMessage(noTimeout)
		if err != nil {
			slog.Error(err.Error())
		}

		if kafkaMessage == nil {
			continue
		}

		// need to experiment with this(when i failed to send email what do i do(mb dlq or just ignore it and let
		// user handler this by calling retry send email himself))
		// for now i will do second cause its easier
		err = c.handler.HandleMessage(kafkaMessage.Value, kafkaMessage.TopicPartition, c.consumerNumber)
		if err != nil {
			// think about dlq
			slog.Error(err.Error())
		}

		_, err = c.consumer.StoreMessage(kafkaMessage)
		if err != nil {
			slog.Error(err.Error())
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

	slog.Info("Commited offset")
	return c.consumer.Close()
}
