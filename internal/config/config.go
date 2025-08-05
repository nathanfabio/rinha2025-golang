package config

import "os"

type Config struct {
	DefaultProcessorURL  string
	FallbackProcessorURL string
	RedisURL             string
	Port                 string
}

func Load() *Config {
	return &Config{
		DefaultProcessorURL:  os.Getenv("PAYMENT_PROCESSOR_URL_DEFAULT"),
		FallbackProcessorURL: os.Getenv("PAYMENT_PROCESSOR_URL_FALLBACK"),
		RedisURL:             os.Getenv("REDIS_URL"),
		Port:                 os.Getenv("APP_PORT"),
	}
}
