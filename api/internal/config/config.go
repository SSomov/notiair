package config

import (
	"os"
	"strconv"
	"strings"
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

type StreamConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

type RedisConfig struct {
	URL string
}

type Config struct {
	HTTP   HTTPConfig
	Queue  QueueConfig
	DB     DatabaseConfig
	Stream StreamConfig
	Redis  RedisConfig
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
		Stream: StreamConfig{
			Brokers: parseBrokers(getEnv("STREAM_BROKERS", "localhost:19092")),
			Topic:   getEnv("STREAM_TOPIC", "test-topic"),
			GroupID: getEnv("STREAM_GROUP_ID", "notiair-workflow-consumer"),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "localhost:6379"),
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

func parseBrokers(brokersStr string) []string {
	if brokersStr == "" {
		return []string{"localhost:19092"}
	}
	brokers := strings.Split(brokersStr, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}
	return brokers
}
