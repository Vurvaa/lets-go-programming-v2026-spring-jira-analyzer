package pusher

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// encapsulates work with the database
type Storage struct {
	db *sql.DB
}

func NewStorage(connectionString string) (*Storage, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("error opening connection to postgres: %w", err)
	}

	db.SetMaxIdleConns(25)
	db.SetMaxOpenConns(25)
	db.SetConnMaxLifetime(3 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging postgres: %w", err)
	}

	return &Storage{db: db}, nil
}

func (stor *Storage) Close() error {
	return stor.db.Close()
}
