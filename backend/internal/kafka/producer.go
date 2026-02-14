package kafka

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	flushTimeout = 5000 // ms
)

var ErrUnknownEventType = errors.New("unknown event type")

type Producer struct {
	producer *kafka.Producer
}

// NewProducer func creates new Producer struct
// need to check if i may need more than one producer
func NewProducer(address []string, passwordSSL, username, passwordUser string) (*Producer, error) {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers":        strings.Join(address, ","),
		"security.protocol":        "SASL_SSL",
		"ssl.ca.location":          "./cert/ca-root.pem",
		"ssl.certificate.location": "./cert/client-certificate.pem",
		"ssl.key.location":         "./cert/client-private-key.pem",
		"ssl.key.password":         passwordSSL,
		"sasl.mechanisms":          "PLAIN",
		"sasl.username":            username,
		"sasl.password":            passwordUser,
	}

	p, err := kafka.NewProducer(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating producer: %s", err.Error())
	}

	return &Producer{
		producer: p,
	}, nil
}

// Produce function produces message to kafka brokers and handles any error that occur during delivery
func (p *Producer) Produce(message interface{}, topic string, key []byte, timestamp time.Time) error {
	js, err := json.MarshalIndent(message, "", "\t")
	if err != nil {
		return fmt.Errorf("error marshaling kafka message into json: %s", err.Error())
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value:     js,
		Key:       key,
		Timestamp: timestamp,
	}

	eventChan := make(chan kafka.Event)

	err = p.producer.Produce(msg, eventChan)
	if err != nil {
		return fmt.Errorf("error producing kafka message: %s", err.Error())
	}

	e := <-eventChan
	switch ev := e.(type) {
	case *kafka.Message:
		return nil
	case kafka.Error:
		return ev
	default:
		return ErrUnknownEventType
	}
}

func (p *Producer) Close() {
	p.producer.Flush(flushTimeout)
	p.producer.Close()
}
