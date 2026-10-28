package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAuthPersistenceAndExpiry(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	hash := uuid.NewString()
	future := time.Now().UTC().Add(time.Hour).Truncate(time.Microsecond)
	attempt := LoginAttempt{StateHash: hash, Verifier: "verifier", Nonce: "nonce", ReturnTo: "/my-servers", ExpiresAt: future}
	if err := s.CreateLoginAttempt(ctx, attempt); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = s.pool.Exec(ctx, `DELETE FROM login_attempts WHERE state_hash = $1`, hash) })
	consumed, err := s.ConsumeLoginAttempt(ctx, hash)
	if err != nil || consumed.Verifier != attempt.Verifier || consumed.Nonce != attempt.Nonce || consumed.ReturnTo != attempt.ReturnTo {
		t.Fatalf("consume = %+v, %v", consumed, err)
	}
	if _, err := s.ConsumeLoginAttempt(ctx, hash); !errors.Is(err, ErrNotFound) {
		t.Fatalf("replayed attempt: %v", err)
	}

	session := Session{Subject: "player", Email: "player@example.com", ExpiresAt: future}
	if err := s.CreateSession(ctx, hash, session); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.DeleteSession(ctx, hash) })
	stored, err := s.GetSession(ctx, hash)
	if err != nil || stored.ID == uuid.Nil || stored.Subject != session.Subject || stored.Email != session.Email || !stored.ExpiresAt.Equal(future) {
		t.Fatalf("session = %+v, %v", stored, err)
	}
	if err := s.DeleteSession(ctx, hash); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetSession(ctx, hash); !errors.Is(err, ErrNotFound) {
		t.Fatalf("revoked session: %v", err)
	}

	attempt.ExpiresAt = time.Now().Add(-time.Hour)
	if err := s.CreateLoginAttempt(ctx, attempt); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ConsumeLoginAttempt(ctx, hash); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expired attempt: %v", err)
	}
	session.ExpiresAt = attempt.ExpiresAt
	if err := s.CreateSession(ctx, hash, session); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetSession(ctx, hash); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expired session: %v", err)
	}
	if err := s.DeleteExpiredAuth(ctx); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM login_attempts WHERE state_hash = $1) + (SELECT count(*) FROM auth_sessions WHERE token_hash = $1)`, hash).Scan(&count); err != nil || count != 0 {
		t.Fatalf("expired rows = %d, %v", count, err)
	}
}

func TestSessionCreatedWithoutEmailRemainsReadable(t *testing.T) {
	s := connectTestStore(t)
	ctx := context.Background()
	hash := uuid.NewString()

	_, err := s.pool.Exec(ctx, `INSERT INTO auth_sessions (token_hash, subject, expires_at) VALUES ($1, $2, $3)`, hash, "legacy-player", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.DeleteSession(ctx, hash) })

	session, err := s.GetSession(ctx, hash)
	if err != nil || session.Subject != "legacy-player" || session.Email != "" {
		t.Fatalf("legacy session = %+v, %v", session, err)
	}
}
