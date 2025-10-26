package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/landly/backend/internal/query"
	_ "github.com/mattn/go-sqlite3"
)

// Config конфигурация SQLite
type Config struct {
	Path            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewConnection создает новое подключение к SQLite
func NewConnection(cfg Config) (*query.Builder, error) {
	db, err := sql.Open("sqlite3", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Проверка подключения
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return query.NewBuilder(query.SQLite, db), nil
}
