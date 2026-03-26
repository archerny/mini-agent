package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/archerny/mini-agent/internal/protocol/uid"
)

// ---------------------------------------------------------------------------
// Message — inter-agent communication unit
// ---------------------------------------------------------------------------

// Message represents a message sent between agents.
type Message struct {
	// ID is a unique identifier (UUID v7, time-ordered).
	ID string `json:"id"`

	// Type is the message type (message / request / response / broadcast).
	Type MessageType `json:"type"`

	// From is the sender agent ID.
	From string `json:"from"`

	// To is the recipient agent ID, or "*" for broadcast.
	To string `json:"to"`

	// CorrelationID links a response to its original request.
	// Only meaningful for agent.response; empty otherwise.
	CorrelationID string `json:"correlation_id,omitempty"`

	// Timestamp is when the message was created.
	Timestamp time.Time `json:"timestamp"`

	// Payload carries the message content.
	Payload Payload `json:"payload"`

	// Metadata is an extensible key-value bag for future use.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Payload is the content section of a message.
type Payload struct {
	// ContentType describes the format of Content.
	ContentType ContentType `json:"content_type"`

	// Content is the actual message content.
	Content string `json:"content"`
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

// NewMessage creates a new Message with a generated ID and current timestamp.
func NewMessage(msgType MessageType, from, to string, payload Payload) *Message {
	return &Message{
		ID:        uid.New(),
		Type:      msgType,
		From:      from,
		To:        to,
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	}
}

// NewRequest creates a new request message.
func NewRequest(from, to string, payload Payload) *Message {
	return NewMessage(TypeRequest, from, to, payload)
}

// NewResponse creates a response message linked to an original request.
func NewResponse(from, to, correlationID string, payload Payload) *Message {
	msg := NewMessage(TypeResponse, from, to, payload)
	msg.CorrelationID = correlationID
	return msg
}

// NewBroadcast creates a broadcast message to all agents.
func NewBroadcast(from string, payload Payload) *Message {
	return NewMessage(TypeBroadcast, from, BroadcastTarget, payload)
}

// TextPayload is a convenience constructor for a text payload.
func TextPayload(content string) Payload {
	return Payload{
		ContentType: ContentText,
		Content:     content,
	}
}

// JSONPayload creates a JSON payload from any value.
func JSONPayload(v any) (Payload, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return Payload{}, fmt.Errorf("marshal json payload: %w", err)
	}
	return Payload{
		ContentType: ContentJSON,
		Content:     string(data),
	}, nil
}

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

// Validate checks the message for structural correctness.
func (m *Message) Validate() error {
	if m.ID == "" {
		return errors.New("message: id is required")
	}
	if m.Type == "" {
		return errors.New("message: type is required")
	}
	if m.From == "" {
		return errors.New("message: from is required")
	}
	if m.To == "" {
		return errors.New("message: to is required")
	}

	// Broadcast must have to = "*"
	if m.Type == TypeBroadcast && m.To != BroadcastTarget {
		return fmt.Errorf("message: broadcast must have to=%q, got %q", BroadcastTarget, m.To)
	}

	// Non-broadcast must not have to = "*"
	if m.Type != TypeBroadcast && m.To == BroadcastTarget {
		return fmt.Errorf("message: non-broadcast must not have to=%q", BroadcastTarget)
	}

	// Response must have correlation_id
	if m.Type == TypeResponse && m.CorrelationID == "" {
		return errors.New("message: response requires correlation_id")
	}

	// Payload size check
	if len(m.Payload.Content) > MaxPayloadSize {
		return fmt.Errorf("message: payload size %d exceeds max %d", len(m.Payload.Content), MaxPayloadSize)
	}

	return nil
}
