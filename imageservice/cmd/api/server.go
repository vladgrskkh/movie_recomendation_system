package main

import (
	"errors"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vladgrskkh/movie-recommender-contracts/v1/imageservice"
	"github.com/vladgrskkh/movie_recomendation_system/imageservice/internal/handlers"
	"google.golang.org/grpc"
)

func server(logger *slog.Logger, address string, h *handlers.ImageHandler) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		logger.Error("failed to listen", slog.String("error", err.Error()))
		os.Exit(1)
	}

	s := grpc.NewServer()
	imageservice.RegisterImageServer(s, h)

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

		signal := <-quit

		logger.Info("shutting down server", slog.String("signal", signal.String()))

		timer := time.AfterFunc(10*time.Second, func() {
			logger.Info("Server couldn't stop gracefully in time. Doing force stop.")
			s.Stop()
		})
		defer timer.Stop()

		s.GracefulStop()
		logger.Info("Server stopped gracefully.")
	}()

	err = s.Serve(lis)
	if !errors.Is(err, grpc.ErrServerStopped) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	logger.Info("server stopped", slog.String("addr", address))

	return nil
}
