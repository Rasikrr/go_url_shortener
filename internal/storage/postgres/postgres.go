package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"go_url_chortener_api/internal/config"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(storageCfg *config.Storage) (*Storage, error) {

	// Name of the function for debugging
	const fn = "storage.postgres.New"

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		storageCfg.Host, storageCfg.Port, storageCfg.User, storageCfg.Dbname,
		storageCfg.SslMode, storageCfg.Password,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", fn, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS url(
    			id SERIAL PRIMARY KEY,
    			alias TEXT NOT NULL UNIQUE,
    			url TEXT NOT NULL);
				CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
			`)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", fn, err)
	}
	return &Storage{
		db: db,
	}, nil
}
