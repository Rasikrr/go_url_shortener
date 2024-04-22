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

func (s *Storage) SaveURL(urlToSave string, alias string) error {
	const fn = "storage.postgres.SaveURL"
	query := `INSERT INTO url(alias, url) VALUES ($1, $2)`
	_, err := s.db.Exec(query, alias, urlToSave)
	if err != nil {
		// TODO unique constraint handling
		return fmt.Errorf("%s: %w", fn, err)
	}
	return nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const fn = "storage.postgres.GetURL"
	query := `SELECT url FROM url WHERE alias=$1 `
	row := s.db.QueryRow(query, alias)
	var url string
	if err := row.Scan(&url); err != nil {
		return "", fmt.Errorf("%s: %w", fn, err)
	}
	return url, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const fn = "storage.postgres.DeleteURL"
	query := `DELETE FROM url WHERE alias=$1`
	_, err := s.db.Exec(query, alias)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}
	return nil
}
