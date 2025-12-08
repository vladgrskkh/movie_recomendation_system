package main

import (
	"log/slog"
	"os"

	"github.com/vladgrskkh/movie_recomendation_system/imageservice/cfg"
)

func main() {
	slog.Info("parsing .env into config")

	cfg, err := cfg.NewConfig()
	if err != nil {
		slog.Error("somethings wrong with .env", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// connection to minio, than gprc server start

}

// TODO: minio repo
// TODO: grpc contract
// TODO: think about how to serve images
