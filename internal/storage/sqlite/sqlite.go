package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"url-shortener/internal/storage"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

// GetURL implements redirect.URLGetter.
func (s *Storage) GetURL(alias string) (string, error) {
	panic("unimplemented")
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New" // operation name for error context

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open sqlite database: %w", op, err)
	}

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url (
		id INTEGER PRIMARY KEY,
		alias TEXT UNIQUE NOT NULL,
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url (alias);
	`)

	if err != nil {
		return nil, fmt.Errorf("%s: failed to prepare create table statement: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute create table statement: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUrl(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveUrl"

	stmt, err := s.db.Prepare("INSERT INTO url (url, alias) VALUES (?, ?)") // #nosec
	if err != nil {
		return 0, fmt.Errorf("%s: failed to prepare insert statement: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		// TODO: refactor this
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetUrlByAlias(alias string) (string, error) {
	const op = "storage.sqlite.GetUrlByAlias" // operapingtion name for error context

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: failed to prepare select statement: %w", op, err)
	}
	defer stmt.Close()

	var resUrl string
	err = stmt.QueryRow(alias).Scan(&resUrl)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}

		return "", fmt.Errorf("%s: failed to execute select statement: %w", op, err)
	}

	return resUrl, nil
}

func (s *Storage) DeleteUrlByAlias(alias string) error {
	const op = "storage.sqlite.DeleteUrlByAlias"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s: failed to prepare delete statement: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: failed to execute delete statement: %w", op, err)
	}

	return nil
}
