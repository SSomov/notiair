package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"notiair/internal/workflow"
	"notiair/services"
)

// Event представляет событие из stream broker
type Event struct {
	EventID    string                 `json:"event_id"`
	EventType  string                 `json:"event_type"`
	OccurredAt string                 `json:"occurred_at"`
	Context    map[string]interface{} `json:"context"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// TriggerConfig представляет конфигурацию Stream broker триггера
type TriggerConfig struct {
	Label       string   `json:"label"`
	Variant     string   `json:"variant"`
	EventTypes  []string `json:"eventTypes"`
	Description string   `json:"description"`
}

// Consumer обрабатывает события из stream broker и запускает workflows
type Consumer struct {
	brokers         []string
	topic           string
	groupID         string
	workflowRepo    workflow.Repository
	notificationSvc *services.NotificationService
	consumer        sarama.ConsumerGroup
	wg              sync.WaitGroup
	hub             *Hub // WebSocket hub для отправки сообщений в интерфейс
	redisStore      *RedisStore // Redis store для хранения сообщений
}

// NewConsumer создает новый consumer для stream broker
func NewConsumer(
	brokers []string,
	topic string,
	groupID string,
	workflowRepo workflow.Repository,
	notificationSvc *services.NotificationService,
	hub *Hub,
	redisStore *RedisStore,
) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Return.Errors = true

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &Consumer{
		brokers:         brokers,
		topic:           topic,
		groupID:         groupID,
		workflowRepo:    workflowRepo,
		notificationSvc: notificationSvc,
		consumer:        consumerGroup,
		hub:             hub,
		redisStore:      redisStore,
	}, nil
}

// Start запускает consumer в фоновом режиме
func (c *Consumer) Start(ctx context.Context) error {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				handler := &consumerGroupHandler{
					consumer:        c,
					workflowRepo:    c.workflowRepo,
					notificationSvc: c.notificationSvc,
				}

				if err := c.consumer.Consume(ctx, []string{c.topic}, handler); err != nil {
					log.Printf("error from consumer: %v", err)
					return
				}
			}
		}
	}()

	// Обработка ошибок consumer
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-c.consumer.Errors():
				if err != nil {
					log.Printf("consumer error: %v", err)
				}
			}
		}
	}()

	log.Printf("stream consumer started for topic: %s, group: %s", c.topic, c.groupID)
	return nil
}

// Stop останавливает consumer
func (c *Consumer) Stop() error {
	if err := c.consumer.Close(); err != nil {
		return fmt.Errorf("failed to close consumer: %w", err)
	}
	c.wg.Wait()
	log.Println("stream consumer stopped")
	return nil
}

// GetRecentMessages получает последние N сообщений из Redis, отфильтрованных по event_types
func GetRecentMessages(redisStore *RedisStore, eventTypes []string, limit int) ([]Event, error) {
	if redisStore == nil {
		return []Event{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return redisStore.GetRecentMessages(ctx, eventTypes, limit)
}

// GetRecentMessagesFromKafka получает последние N сообщений из топика, отфильтрованных по event_types (старый метод, оставлен для совместимости)
func GetRecentMessagesFromKafka(brokers []string, topic string, eventTypes []string, limit int) ([]Event, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}
	defer consumer.Close()

	// Получаем партиции топика
	partitions, err := consumer.Partitions(topic)
	if err != nil {
		return nil, fmt.Errorf("failed to get partitions: %w", err)
	}

	if len(partitions) == 0 {
		return []Event{}, nil
	}

	// Используем первую партицию (обычно 0)
	partitionID := partitions[0]

	// Получаем последний offset
	partitionConsumer, err := consumer.ConsumePartition(topic, partitionID, sarama.OffsetNewest)
	if err != nil {
		return nil, fmt.Errorf("failed to create partition consumer: %w", err)
	}

	lastOffset := partitionConsumer.HighWaterMarkOffset()
	partitionConsumer.Close()

	// Читаем с конца, но ограничиваем количество сообщений для чтения
	readLimit := limit * 10 // Читаем больше, чтобы отфильтровать
	startOffset := lastOffset - int64(readLimit)
	if startOffset < 0 {
		startOffset = 0
	}

	// Создаем новый consumer для чтения с нужного offset
	partitionConsumer, err = consumer.ConsumePartition(topic, partitionID, startOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to create partition consumer: %w", err)
	}
	defer partitionConsumer.Close()

	var events []Event
	eventTypeMap := make(map[string]bool)
	for _, et := range eventTypes {
		eventTypeMap[et] = true
	}

	// Читаем сообщения до последнего offset
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	messageChan := partitionConsumer.Messages()
	for {
		select {
		case <-ctx.Done():
			goto done
		case message, ok := <-messageChan:
			if !ok || message == nil {
				goto done
			}

			if message.Offset >= lastOffset {
				goto done
			}

			var event Event
			if err := json.Unmarshal(message.Value, &event); err != nil {
				continue
			}

			// Фильтруем по event_types, если они указаны
			if len(eventTypes) > 0 && !eventTypeMap[event.EventType] {
				continue
			}

			events = append(events, event)

			// Ограничиваем количество
			if len(events) >= limit {
				goto done
			}
		}
	}

done:
	// Возвращаем в обратном порядке (новые первыми)
	for i, j := 0, len(events)-1; i < j; i, j = i+1, j-1 {
		events[i], events[j] = events[j], events[i]
	}

	return events, nil
}

// consumerGroupHandler реализует sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	consumer        *Consumer
	workflowRepo    workflow.Repository
	notificationSvc *services.NotificationService
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case <-session.Context().Done():
			return nil
		case message := <-claim.Messages():
			if message == nil {
				continue
			}

			var event Event
			if err := json.Unmarshal(message.Value, &event); err != nil {
				log.Printf("failed to unmarshal event: %v", err)
				session.MarkMessage(message, "")
				continue
			}

			log.Printf("received event: %s (type: %s)", event.EventID, event.EventType)

			// Сохраняем сообщение в Redis
			if h.consumer.redisStore != nil {
				if err := h.consumer.redisStore.SaveMessage(session.Context(), event); err != nil {
					log.Printf("failed to save message to redis: %v", err)
				}
			}

			// Отправляем событие в WebSocket hub для интерфейса
			if h.consumer.hub != nil {
				h.consumer.hub.Broadcast(event)
			}

			// Обрабатываем событие
			if err := h.processEvent(session.Context(), event); err != nil {
				log.Printf("failed to process event %s: %v", event.EventID, err)
				// Не коммитим offset при ошибке, чтобы повторить обработку
				continue
			}

			session.MarkMessage(message, "")
		}
	}
}

// processEvent обрабатывает событие и запускает соответствующие workflows
func (h *consumerGroupHandler) processEvent(ctx context.Context, event Event) error {
	// Получаем все активные workflows
	workflows, err := h.workflowRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list workflows: %w", err)
	}

	// Находим workflows с Stream broker триггером, у которых event_type совпадает
	for _, wf := range workflows {
		if !wf.IsActive {
			continue
		}

		// Ищем Stream broker триггер в workflow
		var streamBrokerTrigger *workflow.Node
		var triggerConfig TriggerConfig
		for i := range wf.Nodes {
			node := &wf.Nodes[i]
			if node.Type != workflow.NodeTypeTrigger {
				continue
			}

			// Парсим конфигурацию триггера
			configBytes, err := json.Marshal(node.Config)
			if err != nil {
				continue
			}

			var cfg TriggerConfig
			if err := json.Unmarshal(configBytes, &cfg); err != nil {
				continue
			}

			// Проверяем, что это Stream broker триггер
			if cfg.Label == "Stream broker" {
				streamBrokerTrigger = node
				triggerConfig = cfg
				break
			}
		}

		if streamBrokerTrigger == nil {
			continue
		}

		// Проверяем, подходит ли event_type
		eventTypeMatched := false
		for _, et := range triggerConfig.EventTypes {
			if et == event.EventType {
				eventTypeMatched = true
				break
			}
		}

		if !eventTypeMatched {
			continue
		}

		// Преобразуем событие в payload для workflow
		payload := map[string]interface{}{
			"event_id":    event.EventID,
			"event_type":  event.EventType,
			"occurred_at": event.OccurredAt,
			"context":     event.Context,
			"metadata":    event.Metadata,
		}

		// Запускаем workflow через NotificationService
		// TODO: определить TemplateID и Variables из workflow
		err = h.notificationSvc.Dispatch(ctx, services.DispatchInput{
			WorkflowID: wf.ID,
			TemplateID: "", // Будет определено из workflow
			Variables:  make(map[string]string),
			Payload:    payload,
		})

		if err != nil {
			log.Printf("failed to dispatch workflow %s for event %s: %v", wf.ID, event.EventID, err)
			continue
		}

		log.Printf("dispatched workflow %s for event %s (type: %s)", wf.ID, event.EventID, event.EventType)
	}

	return nil
}

