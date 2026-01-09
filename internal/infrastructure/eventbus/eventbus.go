package eventbus

import (
	"documind.jordi.org/internal/domain/events"
)

// EventBus defines the interface for publishing and subscribing to events
type EventBus interface {
	Publish(event events.Event) error
	Subscribe(eventType events.EventType, handler EventHandler) error
	Start() error
	Stop() error
}

// EventHandler is a function that handles a specific event
type EventHandler func(event events.Event) error
