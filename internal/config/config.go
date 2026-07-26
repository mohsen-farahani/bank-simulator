// Package config provides configuration management for the bank simulator.
// It loads configuration from environment variables with sensible defaults.
package config

import (
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port          string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func Load() *Config {
	cfg := &Config{
		Port:          getEnv("PORT", "8080"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       0,
	}

	// Prefer REDIS_URL when set (Render Key Value, Railway, etc.)
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		if addr, password, db, err := parseRedisURL(redisURL); err == nil {
			cfg.RedisAddr = addr
			cfg.RedisPassword = password
			cfg.RedisDB = db
		}
	}

	return cfg
}

func parseRedisURL(raw string) (addr, password string, db int, err error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", 0, err
	}

	addr = u.Host
	if u.User != nil {
		if p, ok := u.User.Password(); ok {
			password = p
		}
	}

	if path := strings.TrimPrefix(u.Path, "/"); path != "" {
		if n, convErr := strconv.Atoi(path); convErr == nil {
			db = n
		}
	}

	return addr, password, db, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
