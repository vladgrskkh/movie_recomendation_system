package main

import (
	"log/slog"
	"os"
	"os/signal"
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

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("notification server stopped")
	return nil
}

func (app *application) shutdown() error {
	var err error

	for _, c := range app.mailerConsumers {
		closeErr := c.Stop()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}

	return err
}
