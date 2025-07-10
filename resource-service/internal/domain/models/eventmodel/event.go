package eventmodel

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Topic     string    `json:"topic"`
	Payload   []byte    `json:"payload"`
	Sent      bool      `json:"sent"`
	EventTime time.Time `json:"event_time"`
}

func NewEvent[T any](name, topic string, data T) (Event, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return Event{}, err
	}

	return Event{
		Name:    name,
		Topic:   topic,
		Payload: payload,
	}, nil
}

func (e *Event) SetSent() {
	e.Sent = true
}
