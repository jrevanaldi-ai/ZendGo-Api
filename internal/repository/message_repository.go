package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/zendgo/zendgo-api/models"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, message *models.Message) error {
	query := `
		INSERT INTO messages (id, session_id, recipient, message_type, content, media_url, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		message.ID,
		message.SessionID,
		message.Recipient,
		message.MessageType,
		message.Content,
		message.MediaURL,
		message.Status,
		message.CreatedAt,
	)
	return err
}

func (r *MessageRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.MessageStatus, waMessageID string, errorMessage string) error {
	query := `
		UPDATE messages
		SET status = $1, wa_message_id = $2, error_message = $3, sent_at = $4
		WHERE id = $5
	`
	_, err := r.db.ExecContext(ctx, query, status, waMessageID, errorMessage, time.Now(), id)
	return err
}

func (r *MessageRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	query := `
		SELECT id, session_id, recipient, message_type, content, media_url, status, wa_message_id, error_message, created_at, sent_at
		FROM messages
		WHERE id = $1
	`
	message := &models.Message{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&message.ID,
		&message.SessionID,
		&message.Recipient,
		&message.MessageType,
		&message.Content,
		&message.MediaURL,
		&message.Status,
		&message.WAMessageID,
		&message.ErrorMessage,
		&message.CreatedAt,
		&message.SentAt,
	)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (r *MessageRepository) GetBySessionID(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	query := `
		SELECT id, session_id, recipient, message_type, content, media_url, status, wa_message_id, error_message, created_at, sent_at
		FROM messages
		WHERE session_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, sessionID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		message := &models.Message{}
		err := rows.Scan(
			&message.ID,
			&message.SessionID,
			&message.Recipient,
			&message.MessageType,
			&message.Content,
			&message.MediaURL,
			&message.Status,
			&message.WAMessageID,
			&message.ErrorMessage,
			&message.CreatedAt,
			&message.SentAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func (r *MessageRepository) SaveIncomingMessage(ctx context.Context, msg *models.IncomingMessage) error {
	query := `
		INSERT INTO incoming_messages (id, session_id, sender, message_type, content, caption, media_url, timestamp, is_group, from, push_name, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.ExecContext(ctx, query,
		msg.ID,
		msg.SessionID,
		msg.Sender,
		msg.MessageType,
		msg.Content,
		msg.Caption,
		msg.MediaURL,
		msg.Timestamp,
		msg.IsGroup,
		msg.From,
		msg.PushName,
		msg.CreatedAt,
	)
	return err
}
