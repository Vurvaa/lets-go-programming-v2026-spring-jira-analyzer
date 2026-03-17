package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"server/internals/config"

	_ "github.com/lib/pq"
)

func NewDB() *sql.DB {
	cfg := config.LoadDBConfig("configs/config.yaml")

	connectionStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.UserDB,
		cfg.PasswordDB,
		cfg.HostDB,
		cfg.PortDB,
		cfg.NameDB,
	)

	db, err := sql.Open("postgres", connectionStr)
	if err != nil {
		log.Fatalf("Unable to open Postgresql with %s database", connectionStr)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Can't connect to database.")
	}

	return db
}
