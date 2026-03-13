package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type WASessionRepository struct {
	db *sql.DB
}

func NewWASessionRepository(db *sql.DB) *WASessionRepository {
	return &WASessionRepository{db: db}
}

func (r *WASessionRepository) Save(ctx context.Context, sessionID uuid.UUID, identifier string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO wa_sessions (id, session_id, identifier, data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (session_id, identifier) 
		DO UPDATE SET data = $4, updated_at = $6
	`
	_, err = r.db.ExecContext(ctx, query,
		uuid.New(),
		sessionID,
		identifier,
		string(jsonData),
		time.Now(),
		time.Now(),
	)
	return err
}

func (r *WASessionRepository) Get(ctx context.Context, sessionID uuid.UUID, identifier string) ([]byte, error) {
	query := `
		SELECT data
		FROM wa_sessions
		WHERE session_id = $1 AND identifier = $2
	`
	var data string
	err := r.db.QueryRowContext(ctx, query, sessionID, identifier).Scan(&data)
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

func (r *WASessionRepository) GetAll(ctx context.Context, sessionID uuid.UUID) (map[string][]byte, error) {
	query := `
		SELECT identifier, data
		FROM wa_sessions
		WHERE session_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]byte)
	for rows.Next() {
		var identifier string
		var data string
		if err := rows.Scan(&identifier, &data); err != nil {
			return nil, err
		}
		result[identifier] = []byte(data)
	}
	return result, rows.Err()
}

func (r *WASessionRepository) Delete(ctx context.Context, sessionID uuid.UUID, identifier string) error {
	query := `DELETE FROM wa_sessions WHERE session_id = $1 AND identifier = $2`
	_, err := r.db.ExecContext(ctx, query, sessionID, identifier)
	return err
}

func (r *WASessionRepository) DeleteAll(ctx context.Context, sessionID uuid.UUID) error {
	query := `DELETE FROM wa_sessions WHERE session_id = $1`
	_, err := r.db.ExecContext(ctx, query, sessionID)
	return err
}
