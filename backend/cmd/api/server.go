package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) server() error {
	srv := http.Server{
		Addr:         ":8080",
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// gracefull shutdown
	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		// catching signals
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

		s := <-quit

		app.logger.Info("shutting down server", slog.String("signal", s.String()))

		// need to raise timeout to 20 sec (chech if problem solved)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.Info("waiting for background tasks to finish", slog.String("addr", srv.Addr))

		app.wg.Wait()
		shutdownError <- nil
	}()

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// check if something went wrong while shutting down
	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("server stopped", slog.String("addr", srv.Addr))

	return nil
}
