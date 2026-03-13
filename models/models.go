package models

import (
	"time"

	"github.com/google/uuid"
)

type SessionStatus string

const (
	SessionStatusConnecting SessionStatus = "connecting"
	SessionStatusPaired     SessionStatus = "paired"
	SessionStatusDisconnected SessionStatus = "disconnected"
)

type Session struct {
	ID           uuid.UUID     `json:"id" db:"id"`
	Phone        string        `json:"phone" db:"phone"`
	Status       SessionStatus `json:"status" db:"status"`
	WebhookURL   string        `json:"webhook_url,omitempty" db:"webhook_url"`
	APIKey       string        `json:"api_key" db:"api_key"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at" db:"updated_at"`
}

type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeDocument MessageType = "document"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeVideo    MessageType = "video"
	MessageTypeLocation MessageType = "location"
	MessageTypeContact  MessageType = "contact"
)

type MessageStatus string

const (
	MessageStatusQueued    MessageStatus = "queued"
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusFailed    MessageStatus = "failed"
)

type Message struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	SessionID     uuid.UUID     `json:"session_id" db:"session_id"`
	Recipient     string        `json:"recipient" db:"recipient"`
	MessageType   MessageType   `json:"message_type" db:"message_type"`
	Content       string        `json:"content" db:"content"`
	MediaURL      string        `json:"media_url,omitempty" db:"media_url"`
	Status        MessageStatus `json:"status" db:"status"`
	WAMessageID   string        `json:"wa_message_id,omitempty" db:"wa_message_id"`
	ErrorMessage  string        `json:"error_message,omitempty" db:"error_message"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	SentAt        *time.Time    `json:"sent_at,omitempty" db:"sent_at"`
}

type IncomingMessage struct {
	ID          uuid.UUID     `json:"id" db:"id"`
	SessionID   uuid.UUID     `json:"session_id" db:"session_id"`
	Sender      string        `json:"sender" db:"sender"`
	MessageType MessageType   `json:"message_type" db:"message_type"`
	Content     string        `json:"content" db:"content"`
	Caption     string        `json:"caption,omitempty" db:"caption"`
	MediaURL    string        `json:"media_url,omitempty" db:"media_url"`
	Timestamp   int64         `json:"timestamp" db:"timestamp"`
	IsGroup     bool          `json:"is_group" db:"is_group"`
	From        string        `json:"from" db:"from"`
	PushName    string        `json:"push_name,omitempty" db:"push_name"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
}
