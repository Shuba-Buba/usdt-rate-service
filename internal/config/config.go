package config

import (
	"flag"
	"os"
)

type Config struct {
	GRPCPort    string
	PostgresURL string
	GrinexURL   string
	LogLevel    string
}

func New() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.GRPCPort, "grpc-port", "50051", "gRPC server port")
	flag.StringVar(&cfg.PostgresURL, "postgres-url", "postgres://postgres:postgres@localhost:5432/rates?sslmode=disable",
		"PostgreSQL connection URL")
	flag.StringVar(&cfg.GrinexURL, "grinex-url", "https://grinex.io", "Grinex API base URL")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")

	flag.Parse()

	if envPort := os.Getenv("GRPC_PORT"); envPort != "" {
		cfg.GRPCPort = envPort
	}

	if envPostgresURL := os.Getenv("POSTGRES_URL"); envPostgresURL != "" {
		cfg.PostgresURL = envPostgresURL
	}

	if envGrinexURL := os.Getenv("GRINEX_URL"); envGrinexURL != "" {
		cfg.GrinexURL = envGrinexURL
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}

	return cfg
}
