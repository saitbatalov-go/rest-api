package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"rest-api/internal/storage"

	"github.com/glebarez/go-sqlite"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "sqlite.sqlite.New"

	db, err := sql.Open("sqlite", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Остальной код без изменений
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url (
		id INTEGER PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url (alias);`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url (url, alias) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {

		if sqliteErr, ok := err.(*sqlite.Error); ok {
			// Код 2067 - SQLITE_CONSTRAINT_UNIQUE (нарушение уникальности)
			if sqliteErr.Code() == 2067 {
				return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
			}
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(selectedAlias string) (string, error) {
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var resURL string
	err = stmt.QueryRow(selectedAlias).Scan(&resURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resURL, nil
}

func (s *Storage) DeleteURL(selectedAlias string) error {
	const op = "storage.sqlite.DeleteURL"
	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(selectedAlias)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrURLNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
