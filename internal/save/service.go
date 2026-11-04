package save

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/google/uuid"
)

const (
	MaxSavesPerUser = 3
	MaxNameLength   = 40
	Retention       = 90 * 24 * time.Hour
)

var (
	ErrNotFound = errors.New("save not found")
	ErrInUse    = errors.New("a server is still using this save, try again in a few minutes")
	ErrLimit    = errors.New("you can keep at most 3 saves")
)

type InvalidNameError struct {
	Message string
}

func (e InvalidNameError) Error() string {
	return e.Message
}

type saveStore interface {
	CreateSave(ctx context.Context, owner, name string, limit int) (store.SaveRecord, error)
	ListSaves(ctx context.Context, owner string) ([]store.SaveRecord, error)
	RenameSave(ctx context.Context, id uuid.UUID, owner, name string) (store.SaveRecord, error)
	DeleteSave(ctx context.Context, id uuid.UUID, owner string, purge func(context.Context) error) error
	DeleteIdleSave(ctx context.Context, id uuid.UUID, idle time.Duration, purge func(context.Context) error) error
	ListSavesIdleSince(ctx context.Context, idle time.Duration) ([]uuid.UUID, error)
}

type Purger interface {
	Purge(ctx context.Context, prefix string) error
}

type Service struct {
	store  saveStore
	purger Purger
}

func NewService(store saveStore, purger Purger) *Service {
	return &Service{store: store, purger: purger}
}

func Prefix(id uuid.UUID) string {
	return "saves/" + id.String()
}

func normalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > MaxNameLength {
		return "", InvalidNameError{Message: "the save name must be between 1 and 40 characters long"}
	}
	if strings.IndexFunc(name, unicode.IsControl) >= 0 {
		return "", InvalidNameError{Message: "the save name cannot contain control characters"}
	}

	return name, nil
}

func (s *Service) Create(ctx context.Context, owner, name string) (store.SaveRecord, error) {
	name, err := normalizeName(name)
	if err != nil {
		return store.SaveRecord{}, err
	}

	record, err := s.store.CreateSave(ctx, owner, name, MaxSavesPerUser)
	if errors.Is(err, store.ErrSaveLimit) {
		return store.SaveRecord{}, ErrLimit
	}

	return record, err
}

func (s *Service) List(ctx context.Context, owner string) ([]store.SaveRecord, error) {
	return s.store.ListSaves(ctx, owner)
}

func (s *Service) Rename(ctx context.Context, id uuid.UUID, owner, name string) (store.SaveRecord, error) {
	name, err := normalizeName(name)
	if err != nil {
		return store.SaveRecord{}, err
	}

	record, err := s.store.RenameSave(ctx, id, owner, name)
	if errors.Is(err, store.ErrSaveNotFound) {
		return store.SaveRecord{}, ErrNotFound
	}

	return record, err
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, owner string) error {
	err := s.store.DeleteSave(ctx, id, owner, func(ctx context.Context) error { return s.purger.Purge(ctx, Prefix(id)) })
	if errors.Is(err, store.ErrSaveNotFound) {
		return ErrNotFound
	}
	if errors.Is(err, store.ErrSaveInUse) {
		return ErrInUse
	}

	return err
}

func (s *Service) ExpireIdle(ctx context.Context) ([]string, error) {
	ids, err := s.store.ListSavesIdleSince(ctx, Retention)
	if err != nil {
		return nil, err
	}

	expired := make([]string, 0, len(ids))
	var failures []error
	for _, id := range ids {
		err := s.store.DeleteIdleSave(ctx, id, Retention, func(ctx context.Context) error { return s.purger.Purge(ctx, Prefix(id)) })
		if errors.Is(err, store.ErrSaveNotFound) {
			err = ErrNotFound
		}
		if errors.Is(err, store.ErrSaveInUse) {
			err = ErrInUse
		}
		if errors.Is(err, ErrInUse) || errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			slog.Error("Failed to expire save " + id.String() + ": " + err.Error())
			failures = append(failures, err)
			continue
		}
		expired = append(expired, id.String())
	}

	return expired, errors.Join(failures...)
}
