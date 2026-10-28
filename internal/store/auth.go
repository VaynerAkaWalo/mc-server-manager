package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type LoginAttempt struct {
	StateHash string
	Verifier  string
	Nonce     string
	ReturnTo  string
	ExpiresAt time.Time
}

type Session struct {
	ID        uuid.UUID `json:"id"`
	Subject   string    `json:"subject"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func (s *Store) CreateLoginAttempt(ctx context.Context, attempt LoginAttempt) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO login_attempts (state_hash, verifier, nonce, return_to, expires_at) VALUES ($1, $2, $3, $4, $5)`,
		attempt.StateHash, attempt.Verifier, attempt.Nonce, attempt.ReturnTo, attempt.ExpiresAt)
	return err
}

func (s *Store) ConsumeLoginAttempt(ctx context.Context, hash string) (LoginAttempt, error) {
	var attempt LoginAttempt
	err := s.pool.QueryRow(ctx, `DELETE FROM login_attempts WHERE state_hash = $1 AND expires_at > now() RETURNING verifier, nonce, return_to, expires_at`, hash).
		Scan(&attempt.Verifier, &attempt.Nonce, &attempt.ReturnTo, &attempt.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return LoginAttempt{}, ErrNotFound
	}
	return attempt, err
}

func (s *Store) CreateSession(ctx context.Context, hash string, session Session) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO auth_sessions (token_hash, subject, expires_at) VALUES ($1, $2, $3)`, hash, session.Subject, session.ExpiresAt)
	return err
}

func (s *Store) GetSession(ctx context.Context, hash string) (Session, error) {
	var session Session
	err := s.pool.QueryRow(ctx, `SELECT id, subject, expires_at FROM auth_sessions WHERE token_hash = $1 AND expires_at > now()`, hash).
		Scan(&session.ID, &session.Subject, &session.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	return session, err
}

func (s *Store) DeleteSession(ctx context.Context, hash string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM auth_sessions WHERE token_hash = $1`, hash)
	return err
}

func (s *Store) DeleteExpiredAuth(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, `DELETE FROM login_attempts WHERE expires_at <= now()`); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `DELETE FROM auth_sessions WHERE expires_at <= now()`)
	return err
}
