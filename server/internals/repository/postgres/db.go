package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"server/internals/config"

	_ "github.com/lib/pq"
)

func NewDB(configName string) *sql.DB {
	cfg := config.LoadDBConfig(configName)
	if cfg == nil {
		log.Fatalf("Unable to load db config %s", configName)
	}

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
