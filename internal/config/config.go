// Package config provides configuration management for the bank simulator.
// It loads configuration from environment variables with sensible defaults.
package config

import (
	"os"
)

type Config struct {
	Port          string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "80"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", "1234"),
		RedisDB:       0,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
