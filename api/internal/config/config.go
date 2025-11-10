package config

import (
	"os"
	"strconv"
)

type HTTPConfig struct {
	Addr string
}

type QueueConfig struct {
	URL        string
	Namespace  string
	RetryLimit int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type Config struct {
	HTTP  HTTPConfig
	Queue QueueConfig
	DB    DatabaseConfig
}

func Load() (Config, error) {
	cfg := Config{
		HTTP: HTTPConfig{
			Addr: getEnv("HTTP_ADDR", ":8080"),
		},
		Queue: QueueConfig{
			URL:        getEnv("QUEUE_URL", "redis://localhost:6379"),
			Namespace:  getEnv("QUEUE_NAMESPACE", "notiair"),
			RetryLimit: getEnvInt("QUEUE_RETRY_LIMIT", 5),
		},
		DB: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "notiair"),
			Password: getEnv("DB_PASSWORD", "notiair"),
			Name:     getEnv("DB_NAME", "notiair"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return fallback
}
