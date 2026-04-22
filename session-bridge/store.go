package main

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct{ db *sql.DB }

func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS codes (
        code TEXT PRIMARY KEY,
        sub TEXT NOT NULL,
        expires_at INTEGER NOT NULL,
        used INTEGER NOT NULL DEFAULT 0
    )`)
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Issue(sub string, ttl time.Duration) (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(100_000_000))
	if err != nil {
		return "", err
	}
	code := fmt.Sprintf("%08d", n.Int64())
	exp := time.Now().Add(ttl).Unix()
	_, err = s.db.Exec("INSERT INTO codes(code, sub, expires_at) VALUES(?,?,?)", code, sub, exp)
	return code, err
}

var ErrInvalidCode = errors.New("invalid code")

func (s *Store) Claim(code string) (string, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var sub string
	var exp int64
	var used int
	err = tx.QueryRow("SELECT sub, expires_at, used FROM codes WHERE code=?", code).
		Scan(&sub, &exp, &used)
	if err != nil {
		return "", ErrInvalidCode
	}
	if used == 1 || time.Now().Unix() > exp {
		return "", ErrInvalidCode
	}

	if _, err := tx.Exec("UPDATE codes SET used=1 WHERE code=?", code); err != nil {
		return "", err
	}
	return sub, tx.Commit()
}
