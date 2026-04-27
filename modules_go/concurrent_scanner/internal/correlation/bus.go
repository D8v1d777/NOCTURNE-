package events

import (
	"sync"
	"time"
)

// Topic defines the streaming channels
type Topic string

const (
	TopicRawData            Topic = "raw_data"
	TopicNormalizedIdentity Topic = "normalized_identities"
	TopicCorrelationEvent   Topic = "correlation_events"
	TopicAlerts             Topic = "alerts"
)

// Event represents the message payload
type Event struct {
	Topic     Topic
	TargetID  string
	Payload   interface{}
	Timestamp time.Time
}

// StreamBus handles the pub/sub logic
type StreamBus struct {
	mu          sync.RWMutex
	subscribers map[Topic][]chan Event
	closed      bool
}

func NewStreamBus() *StreamBus {
	return &StreamBus{
		subscribers: make(map[Topic][]chan Event),
	}
}

// Subscribe creates a new consumer channel for a specific topic
func (b *StreamBus) Subscribe(topic Topic) <-chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan Event, 100)
	b.subscribers[topic] = append(b.subscribers[topic], ch)
	return ch
}

// Publish sends an event to all subscribers of a topic
func (b *StreamBus) Publish(topic Topic, targetID string, payload interface{}) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	event := Event{
		Topic:     topic,
		TargetID:  targetID,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}

	for _, ch := range b.subscribers[topic] {
		// Non-blocking publish to avoid slow consumers stalling the pipeline
		select {
		case ch <- event:
		default:
		}
	}
}

func (b *StreamBus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
}
