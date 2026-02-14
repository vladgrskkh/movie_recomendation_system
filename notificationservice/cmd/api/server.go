package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func (app *application) server() error {
	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

		s := <-quit

		app.logger.Info("shutting down notification server", slog.String("signal", s.String()))

		shutdownError <- app.shutdown()
	}()

	err := app.startMailerConsumers()
	if err != nil {
		return err
	}

	err = app.startDummyKafkaConsumers()
	if err != nil {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("notification server stopped")
	return nil
}

func (app *application) shutdown() error {
	var err error
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		for _, c := range app.mailerConsumers {
			closeErr := c.Stop()
			if closeErr != nil && err == nil {
				app.logger.Error("something went wrong while shutting down consumers", slog.String("error", closeErr.Error()))
				err = fmt.Errorf("%w, %w", err, closeErr)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for _, c := range app.dummyConsumers {
			closeErr := c.Stop()
			if closeErr != nil && err == nil {
				app.logger.Error("something went wrong while shutting down consumers", slog.String("error", closeErr.Error()))
				err = fmt.Errorf("%w, %w", err, closeErr)
			}
		}
	}()

	wg.Wait()

	return err
}
