package config

import "os"

type Config struct {
	DatabaseURL string
	RedisAddr   string
	KafkaBrokers string
	HTTPPort    string
}

func Load() Config {
	return Config{
		DatabaseURL:  env("DATABASE_URL", "postgres://app:app@localhost:5432/orders?sslmode=disable"),
		RedisAddr:    env("REDIS_ADDR", "localhost:6379"),
		KafkaBrokers: env("KAFKA_BROKERS", "localhost:19092"),
		HTTPPort:     env("HTTP_PORT", "8080"),
	}
}

func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
