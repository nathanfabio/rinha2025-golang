package config

import (
	"os"
	"time"
)

type Config struct {
	DefaultProcessorURL  string
	FallbackProcessorURL string
	RedisURL             string
	Port                 string
	HealthCheckInterval  time.Duration
}

func Load() *Config {
	return &Config{
		DefaultProcessorURL:  os.Getenv("PAYMENT_PROCESSOR_URL_DEFAULT"),
		FallbackProcessorURL: os.Getenv("PAYMENT_PROCESSOR_URL_FALLBACK"),
		RedisURL:             os.Getenv("REDIS_URL"),
		Port:                 os.Getenv("APP_PORT"),
		HealthCheckInterval:  5 * time.Second,
	}
}
