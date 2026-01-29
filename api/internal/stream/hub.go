package stream

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

// Hub управляет WebSocket соединениями и рассылкой сообщений
type Hub struct {
	clients    map[*websocket.Conn]map[string]bool // client -> eventTypes filter
	broadcast  chan Event
	register   chan *Client
	unregister chan *websocket.Conn
	mu         sync.RWMutex
	redisStore *RedisStore // Redis store для сохранения сообщений
}

// Client представляет WebSocket клиента с фильтром event types
type Client struct {
	conn       *websocket.Conn
	eventTypes map[string]bool
}

// NewHub создает новый hub для WebSocket соединений
func NewHub(redisStore *RedisStore) *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]map[string]bool),
		broadcast:  make(chan Event, 256),
		register:   make(chan *Client),
		unregister: make(chan *websocket.Conn),
		redisStore: redisStore,
	}
}

// Run запускает hub для обработки соединений и рассылки сообщений
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.conn] = client.eventTypes
			h.mu.Unlock()
			log.Printf("WebSocket client connected. Total clients: %d", len(h.clients))

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
			h.mu.Unlock()
			log.Printf("WebSocket client disconnected. Total clients: %d", len(h.clients))

		case event := <-h.broadcast:
			// Сохраняем сообщение в Redis
			if h.redisStore != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				if err := h.redisStore.SaveMessage(ctx, event); err != nil {
					log.Printf("failed to save message to redis in hub: %v", err)
				}
				cancel()
			}

			h.mu.RLock()
			for conn, eventTypes := range h.clients {
				// Если у клиента нет фильтров или event_type в фильтрах, отправляем сообщение
				if len(eventTypes) == 0 || eventTypes[event.EventType] {
					data, err := json.Marshal(event)
					if err != nil {
						log.Printf("failed to marshal event: %v", err)
						continue
					}

					if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
						log.Printf("failed to write message: %v", err)
						h.unregister <- conn
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast отправляет событие всем подключенным клиентам
func (h *Hub) Broadcast(event Event) {
	select {
	case h.broadcast <- event:
	default:
		log.Println("broadcast channel full, dropping event")
	}
}

// RegisterClient регистрирует нового клиента с фильтром event types
func (h *Hub) RegisterClient(conn *websocket.Conn, eventTypes []string) {
	eventTypeMap := make(map[string]bool)
	for _, et := range eventTypes {
		eventTypeMap[et] = true
	}

	client := &Client{
		conn:       conn,
		eventTypes: eventTypeMap,
	}

	h.register <- client
}

// UnregisterClient отменяет регистрацию клиента
func (h *Hub) UnregisterClient(conn *websocket.Conn) {
	h.unregister <- conn
}

