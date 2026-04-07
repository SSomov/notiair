package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// Redis key prefix для хранения сообщений по event types
	redisKeyPrefix = "stream:messages:"
	// Максимальное количество сообщений для каждого event type
	maxMessagesPerType = 10
)

// RedisStore управляет хранением сообщений в Redis
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore создает новый Redis store
func NewRedisStore(redisURL string) (*RedisStore, error) {
	var opt *redis.Options
	
	// Пробуем распарсить как URL (redis://localhost:6379)
	if strings.HasPrefix(redisURL, "redis://") || strings.HasPrefix(redisURL, "rediss://") {
		parsedOpt, err := redis.ParseURL(redisURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse redis URL: %w", err)
		}
		opt = parsedOpt
	} else {
		// Если не URL, используем как адрес (localhost:6379)
		opt = &redis.Options{
			Addr: redisURL,
		}
	}

	client := redis.NewClient(opt)

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisStore{client: client}, nil
}

// SaveMessage сохраняет сообщение в Redis для соответствующего event type
func (r *RedisStore) SaveMessage(ctx context.Context, event Event) error {
	key := redisKeyPrefix + event.EventType

	// Сериализуем событие в JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Добавляем сообщение в начало списка (LPUSH)
	pipe := r.client.Pipeline()
	pipe.LPush(ctx, key, data)
	pipe.LTrim(ctx, key, 0, maxMessagesPerType-1) // Оставляем только последние 10
	pipe.Expire(ctx, key, 24*time.Hour)           // Устанавливаем TTL 24 часа

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save message to redis: %w", err)
	}

	return nil
}

// GetRecentMessages получает последние N сообщений для указанных event types
func (r *RedisStore) GetRecentMessages(ctx context.Context, eventTypes []string, limit int) ([]Event, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	var allEvents []Event

	if len(eventTypes) == 0 {
		// Если event types не указаны, получаем все ключи с префиксом
		keys, err := r.client.Keys(ctx, redisKeyPrefix+"*").Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get keys: %w", err)
		}

		// Собираем сообщения из всех ключей
		for _, key := range keys {
			messages, err := r.getMessagesFromKey(ctx, key, limit)
			if err != nil {
				continue // Пропускаем ошибки для отдельных ключей
			}
			allEvents = append(allEvents, messages...)
		}
	} else {
		// Получаем сообщения только для указанных event types
		for _, eventType := range eventTypes {
			key := redisKeyPrefix + eventType
			messages, err := r.getMessagesFromKey(ctx, key, limit)
			if err != nil {
				continue // Пропускаем ошибки для отдельных ключей
			}
			allEvents = append(allEvents, messages...)
		}
	}

	// Сортируем по времени (новые первыми) и ограничиваем количество
	// События уже отсортированы в Redis (новые первыми), просто берем первые limit
	if len(allEvents) > limit {
		allEvents = allEvents[:limit]
	}

	return allEvents, nil
}

// getMessagesFromKey получает сообщения из конкретного ключа Redis
func (r *RedisStore) getMessagesFromKey(ctx context.Context, key string, limit int) ([]Event, error) {
	// Получаем последние N сообщений из списка
	data, err := r.client.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	var events []Event
	for _, item := range data {
		var event Event
		if err := json.Unmarshal([]byte(item), &event); err != nil {
			continue // Пропускаем некорректные сообщения
		}
		events = append(events, event)
	}

	return events, nil
}

// Close закрывает соединение с Redis
func (r *RedisStore) Close() error {
	return r.client.Close()
}

