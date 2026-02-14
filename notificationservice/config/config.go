package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
)

type Config struct {
	Mailer struct {
		MailerAPIKey string
		Sender       string
	}
	ConsumerMailer struct {
		ConsumerGroup string `toml:"consumer_group"`
		ConsumerCount int    `toml:"consumer_count"`
		Topic         string `toml:"topic"`
	} `toml:"consumer_mailer"`
}

func New() (*Config, error) {
	var cfgFile string
	var cfg Config

	flag.StringVar(&cfgFile, "cfg", "config.toml", "TOML config file path(vary based on where deploy)")

	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env: %w", err)
	}

	cfg.Mailer.MailerAPIKey = os.Getenv("MAILERSEND_API_KEY")
	cfg.Mailer.Sender = os.Getenv("SMTP_USERNAME")

	metadata, err := toml.DecodeFile(cfgFile, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error loading .toml config file: %w", err)
	}

	if len(metadata.Undecoded()) > 0 {
		return nil, fmt.Errorf("unknown configuration keys: %v", metadata.Undecoded())
	}

	return &cfg, nil
}
