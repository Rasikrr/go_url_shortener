package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"go_url_chortener_api/internal/config"
	"go_url_chortener_api/internal/domain"
	"go_url_chortener_api/internal/http-server/handlers/refresh"
	"go_url_chortener_api/internal/storage"
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

	if err := createTables(db); err != nil {
		return nil, err
	}

	return &Storage{
		db: db,
	}, nil
}

func createTables(db *sql.DB) error {
	const fn = "storage.postgres.createTables"

	_, err := db.Exec(
		`CREATE TABLE IF NOT EXISTS url(
    			id SERIAL PRIMARY KEY,
    			alias TEXT NOT NULL UNIQUE,
    			url TEXT NOT NULL);
				CREATE INDEX IF NOT EXISTS idx_alias ON url(alias
				    );
			`)
	if err != nil {
		return fmt.Errorf("%s : %w", fn, err)
	}

	_, err = db.Exec(`
				CREATE TABLE IF NOT EXISTS users(
				    id SERIAL PRIMARY KEY,
				    email VARCHAR(50) NOT NULL UNIQUE,
				    enc_password VARCHAR(256)
				);
			`)
	if err != nil {
		return fmt.Errorf("%s : %w", fn, err)
	}

	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS refresh_token(
    				id SERIAL PRIMARY KEY,
    				token VARCHAR(256),
    				user_id INT,
    				CONSTRAINT refresh_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
				);
			`)
	if err != nil {
		return fmt.Errorf("%s : %w", fn, err)
	}
	return nil
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
	res, err := s.db.Exec(query, alias)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	deleted, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	if deleted < 1 {
		return storage.ErrURLNotFound
	}
	return nil
}

func (s *Storage) GetUser(email string) (*domain.User, error) {
	const fn = "storage.postgres.GetUser"

	query := `SELECT * FROM users WHERE email=$1`
	row := s.db.QueryRow(query, email)

	user := new(domain.User)
	err := row.Scan(&user.Id, &user.Email, &user.EncPassword)

	if err != nil {
		return nil, fmt.Errorf("%s : %w", fn, err)
	}
	return user, nil
}

func (s *Storage) GetUserById(id int) (*domain.User, error) {
	const fn = "storage.postgres.GetUserById"

	query := `SELECT * FROM users WHERE id=$1`

	user := new(domain.User)
	row := s.db.QueryRow(query, id)

	err := row.Scan(&user.Id, &user.Email, &user.EncPassword)

	if err != nil {
		return nil, fmt.Errorf("%s : %w", fn, err)
	}
	return user, nil
}

func (s *Storage) SaveUser(user *domain.User) error {
	const fn = "storage.postgres.SaveUser"
	query := `INSERT INTO users(email, enc_password)
				VALUES($1, $2);`

	if _, err := s.db.Exec(query, user.Email, user.EncPassword); err != nil {
		return fmt.Errorf("%s : %w", err)
	}

	return nil
}

func (s *Storage) SaveRefresh(token string, userId int) error {
	const fn = "storage.postgres.SaveRefresh"

	query := `INSERT INTO refresh_token (token, user_id)
				VALUES ($1, $2)`

	if _, err := s.db.Exec(query, token, userId); err != nil {
		return fmt.Errorf("%s : %w", fn, err)
	}
	return nil

}

func (s *Storage) GetRefresh(userId int) (*refresh.Token, error) {
	const fn = "storage.postgres.GetRefresh"

	query := `SELECT * FROM refresh_token WHERE user_id=$1`

	row := s.db.QueryRow(query, userId)

	token := new(refresh.Token)

	if err := row.Scan(&token.Id, &token.Token, &token.UserId); err != nil {
		return nil, fmt.Errorf("%s : %w", fn, err)
	}
	return token, nil
}

func (s *Storage) DeleteRefresh(token string) error {
	const fn = "storage.postgres.DeleteRefresh"
	query := `DELETE FROM refresh_token WHERE token=$1`
	if _, err := s.db.Exec(query, token); err != nil {
		return fmt.Errorf("%s : %w", fn, err)
	}
	return nil
}

func (s *Storage) DeleteRefreshByUserId(id int) error {
	const fn = "storage.postgres.DeleteRefreshByUserId"
	query := `DELETE FROM refresh_token WHERE user_id=$1`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("%s : %w", fn, err)
	}
	return nil
}

func (s *Storage) UpdateRefresh(id int, token string) error {
	const fn = "storage.postgres.UpdateRefresh"
	query := `UPDATE refresh_token SET token=$1 WHERE id=$2`
	if _, err := s.db.Exec(query, token, id); err != nil {
		return fmt.Errorf("%s : %w", err)
	}
	return nil

}
