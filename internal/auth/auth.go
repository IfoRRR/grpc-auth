package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("username taken")
	ErrUnauthorized       = errors.New("unauthorized")
)

// Service — JWT auth; users in Postgres.
type Service struct {
	secret []byte
	db     *sql.DB
}

func New(secret string, db *sql.DB) (*Service, error) {
	if secret == "" {
		return nil, errors.New("jwt secret required")
	}
	if db == nil {
		return nil, errors.New("db required")
	}
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS users (
    username TEXT PRIMARY KEY,
    password_hash BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`)
	if err != nil {
		return nil, err
	}
	return &Service{secret: []byte(secret), db: db}, nil
}

func (s *Service) Register(username, password string) (string, error) {
	username = strings.TrimSpace(username)
	if len(username) < 2 || len(username) > 32 || len(password) < 8 || len(password) > 128 {
		return "", fmt.Errorf("username 2-32 chars, password 8-128 chars")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	_, err = s.db.Exec(`INSERT INTO users (username, password_hash) VALUES ($1, $2)`, username, hash)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return "", ErrUserExists
		}
		return "", err
	}
	return s.issue(username)
}

func (s *Service) Login(username, password string) (string, error) {
	var hash []byte
	err := s.db.QueryRow(
		`SELECT password_hash FROM users WHERE username = $1`,
		strings.TrimSpace(username),
	).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrInvalidCredentials
	}
	if err != nil {
		return "", err
	}
	if bcrypt.CompareHashAndPassword(hash, []byte(password)) != nil {
		return "", ErrInvalidCredentials
	}
	return s.issue(username)
}

func (s *Service) Validate(token string) (string, error) {
	token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
	if token == "" {
		return "", ErrUnauthorized
	}
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil || !parsed.Valid {
		return "", ErrUnauthorized
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrUnauthorized
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", ErrUnauthorized
	}
	return sub, nil
}

func (s *Service) issue(username string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})
	return t.SignedString(s.secret)
}
