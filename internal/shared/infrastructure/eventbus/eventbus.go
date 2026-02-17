package eventbus

import (
	"documind.jordi.org/internal/shared/domain"
)

// EventBus defines the interface for publishing and subscribing to events
type EventBus interface {
	Publish(event domain.Event) error
	Subscribe(eventType domain.EventType, handler EventHandler) error
	Start() error
	Stop() error
}

// EventHandler is a function that handles a specific event
type EventHandler func(event domain.Event) error
