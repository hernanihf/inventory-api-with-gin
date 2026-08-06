package database

import (
	"database/sql"
	"fmt"
	"inventory_api/config"
	"os"

	_ "github.com/lib/pq"
)

func ConnectToDatabase(cfg *config.Config) (*sql.DB, error) {
	host := cfg.Database.Host
	if envHost := os.Getenv("DB_HOST"); envHost != "" {
		host = envHost
	}

	url := fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%d sslmode=disable",
		cfg.Database.Username, cfg.Database.Dbname, cfg.Database.Password, host, cfg.Database.Port)

	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
