package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/zendgo/zendgo-api/models"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (id, phone, status, webhook_url, api_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.Phone,
		session.Status,
		session.WebhookURL,
		session.APIKey,
		session.CreatedAt,
		session.UpdatedAt,
	)
	return err
}

func (r *SessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	query := `
		SELECT id, phone, status, webhook_url, api_key, created_at, updated_at
		FROM sessions
		WHERE id = $1
	`
	session := &models.Session{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.Phone,
		&session.Status,
		&session.WebhookURL,
		&session.APIKey,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (r *SessionRepository) GetByAPIKey(ctx context.Context, apiKey string) (*models.Session, error) {
	query := `
		SELECT id, phone, status, webhook_url, api_key, created_at, updated_at
		FROM sessions
		WHERE api_key = $1
	`
	session := &models.Session{}
	err := r.db.QueryRowContext(ctx, query, apiKey).Scan(
		&session.ID,
		&session.Phone,
		&session.Status,
		&session.WebhookURL,
		&session.APIKey,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (r *SessionRepository) GetByPhone(ctx context.Context, phone string) (*models.Session, error) {
	query := `
		SELECT id, phone, status, webhook_url, api_key, created_at, updated_at
		FROM sessions
		WHERE phone = $1
	`
	session := &models.Session{}
	err := r.db.QueryRowContext(ctx, query, phone).Scan(
		&session.ID,
		&session.Phone,
		&session.Status,
		&session.WebhookURL,
		&session.APIKey,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (r *SessionRepository) GetAll(ctx context.Context) ([]*models.Session, error) {
	query := `
		SELECT id, phone, status, webhook_url, api_key, created_at, updated_at
		FROM sessions
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		session := &models.Session{}
		err := rows.Scan(
			&session.ID,
			&session.Phone,
			&session.Status,
			&session.WebhookURL,
			&session.APIKey,
			&session.CreatedAt,
			&session.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

func (r *SessionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.SessionStatus) error {
	query := `
		UPDATE sessions
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

func (r *SessionRepository) UpdateWebhookURL(ctx context.Context, id uuid.UUID, webhookURL string) error {
	query := `
		UPDATE sessions
		SET webhook_url = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, webhookURL, time.Now(), id)
	return err
}

func (r *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sessions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
