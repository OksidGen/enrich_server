package config

import (
	"fmt"
	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type (
	Config struct {
		PG `envPrefix:"PG_"`
	}

	PG struct {
		USER     string `env:"USER"`
		PASSWORD string `env:"PASSWORD"`
		HOST     string `env:"HOST"`
		PORT     int    `env:"PORT"`
		DATABASE string `env:"DATABASE"`
	}
)

func NewConfig() (*Config, error) {
	loadEnv()
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse env: %w", err)
	}
	return &cfg, nil
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Err(err).Msg("Failed to load .env file")
	}
}
