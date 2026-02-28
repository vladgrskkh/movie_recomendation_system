package cfg

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type minioConfig struct {
	Endpoint   string `env:"MINIO_ENDPOINT"`
	AccessKey  string `env:"MINIO_ACCESS_KEY"`
	SecretKey  string `env:"MINIO_SECRET_KEY"`
	BucketName string `env:"MINIO_BUCKET_NAME"`
}
type Config struct {
	Port    string `env:"PORT"`
	Address string `env:"ADDRESS"`
	Minio   minioConfig
}

func NewConfig() (*Config, error) {
	var config Config

	err := env.Parse(&config)
	if err != nil {
		return nil, fmt.Errorf("error parsing .env into config: %w", err)
	}

	return &config, nil
}
