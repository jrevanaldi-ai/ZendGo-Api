package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/zendgo/zendgo-api/internal/config"
	"github.com/zendgo/zendgo-api/models"
)

type Database struct {
	DB       *sql.DB
	Config   *config.DatabaseConfig
}

func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return &Database{
		DB:     db,
		Config: cfg,
	}, nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}

func (d *Database) Migrate() error {
	ctx := context.Background()
	migrations := models.GetAllMigrations()

	for _, migration := range migrations {
		log.Printf("Running migration: %s", migration.Name)
		if _, err := d.DB.ExecContext(ctx, migration.Query); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.Name, err)
		}
		log.Printf("Migration %s completed", migration.Name)
	}

	log.Println("All migrations completed successfully")
	return nil
}
