package eventbus

import (
	"documind.jordi.org/internal/shared/domain"
	"fmt"
	"log"
	"sync"
)

// MemoryBus is an in-memory event bus for development
type MemoryBus struct {
	handlers map[domain.EventType][]EventHandler
	mu       sync.RWMutex
}

func NewMemoryBus() *MemoryBus {
	return &MemoryBus{
		handlers: make(map[domain.EventType][]EventHandler),
	}
}

func (b *MemoryBus) Publish(event domain.Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	handlers, exists := b.handlers[event.GetType()]
	if !exists {
		log.Printf("No handlers registered for event type: %s", event.GetType())
		return nil
	}

	for _, handler := range handlers {
		// Execute handlers synchronously in development
		if err := handler(event); err != nil {
			return fmt.Errorf("handler error for event %s: %w", event.GetType(), err)
		}
	}

	return nil
}

func (b *MemoryBus) Subscribe(eventType domain.EventType, handler EventHandler) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
	log.Printf("Subscribed handler for event type: %s", eventType)
	return nil
}

func (b *MemoryBus) Start() error {
	log.Println("MemoryBus started")
	return nil
}

func (b *MemoryBus) Stop() error {
	log.Println("MemoryBus stopped")
	return nil
}
