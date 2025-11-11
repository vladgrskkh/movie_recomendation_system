package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/vladgrskkh/movie_recomendation_system/notificationservice/internal/consumer"
)

func (app *application) startMailerConsumers() error {
	handler := app.NewSendEmailHandler()
	consumers := make([]*consumer.Consumer, app.config.ConsumerMailer.ConsumerCount)
	for i := 1; i <= app.config.ConsumerMailer.ConsumerCount; i++ {
		c, err := consumer.NewConsumer(handler, app.config.Address, app.config.ConsumerMailer.Topic, app.config.ConsumerMailer.ConsumerGroup, i)
		if err != nil {
			return fmt.Errorf("failed to start consumer %d, error detail: %s", i, err.Error())
		}

		go c.Start()
		consumers = append(consumers, c)
	}

	app.mailerConsumers = consumers
	return nil
}

func readJSON(body []byte, dst interface{}) error {
	byteReader := bytes.NewBuffer(body)
	dec := json.NewDecoder(byteReader)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formated JSON(at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formated JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type(at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}
