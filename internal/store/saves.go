package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrSaveLimit = errors.New("the save limit is reached")

type SaveRecord struct {
	ID        uuid.UUID
	Owner     string
	Name      string
	CreatedAt time.Time
}

func lockSave(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, "save:"+id.String()); err != nil {
		return fmt.Errorf("lock save: %w", err)
	}

	return nil
}

func (s *Store) CreateSave(ctx context.Context, owner, name string, limit int) (SaveRecord, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return SaveRecord{}, fmt.Errorf("begin save insert: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, "saves:"+owner); err != nil {
		return SaveRecord{}, fmt.Errorf("lock saves of owner: %w", err)
	}

	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM saves WHERE owner_subject = $1`, owner).Scan(&count); err != nil {
		return SaveRecord{}, fmt.Errorf("count saves: %w", err)
	}
	if count >= limit {
		return SaveRecord{}, ErrSaveLimit
	}

	record := SaveRecord{Owner: owner, Name: name}
	err = tx.QueryRow(ctx, `INSERT INTO saves (owner_subject, name) VALUES ($1, $2) RETURNING id, created_at`, owner, name).Scan(&record.ID, &record.CreatedAt)
	if err != nil {
		return SaveRecord{}, fmt.Errorf("insert save: %w", err)
	}

	return record, tx.Commit(ctx)
}

func (s *Store) ListSaves(ctx context.Context, owner string) ([]SaveRecord, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, owner_subject, name, created_at FROM saves WHERE owner_subject = $1 ORDER BY created_at, id`, owner)
	if err != nil {
		return nil, fmt.Errorf("list saves: %w", err)
	}
	defer rows.Close()

	records := make([]SaveRecord, 0)
	for rows.Next() {
		var record SaveRecord
		if err := rows.Scan(&record.ID, &record.Owner, &record.Name, &record.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan save: %w", err)
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (s *Store) GetSave(ctx context.Context, id uuid.UUID, owner string) (SaveRecord, error) {
	var record SaveRecord
	err := s.pool.QueryRow(ctx, `SELECT id, owner_subject, name, created_at FROM saves WHERE id = $1 AND owner_subject = $2`, id, owner).
		Scan(&record.ID, &record.Owner, &record.Name, &record.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return SaveRecord{}, ErrSaveNotFound
	}
	if err != nil {
		return SaveRecord{}, fmt.Errorf("get save: %w", err)
	}

	return record, nil
}

func (s *Store) RenameSave(ctx context.Context, id uuid.UUID, owner, name string) (SaveRecord, error) {
	var record SaveRecord
	err := s.pool.QueryRow(ctx, `UPDATE saves SET name = $3 WHERE id = $1 AND owner_subject = $2 RETURNING id, owner_subject, name, created_at`, id, owner, name).
		Scan(&record.ID, &record.Owner, &record.Name, &record.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return SaveRecord{}, ErrSaveNotFound
	}
	if err != nil {
		return SaveRecord{}, fmt.Errorf("rename save: %w", err)
	}

	return record, nil
}

func (s *Store) DeleteSave(ctx context.Context, id uuid.UUID, owner string, purge func(context.Context) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin save delete: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := lockSave(ctx, tx, id); err != nil {
		return err
	}

	var held bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM mc_servers WHERE save_id = $1 AND save_locked)`, id).Scan(&held); err != nil {
		return fmt.Errorf("check save lock: %w", err)
	}
	if held {
		return ErrSaveInUse
	}

	found, err := tx.Exec(ctx, `SELECT 1 FROM saves WHERE id = $1 AND ($2 = '' OR owner_subject = $2)`, id, owner)
	if err != nil {
		return fmt.Errorf("find save: %w", err)
	}
	if found.RowsAffected() == 0 {
		return ErrSaveNotFound
	}

	if err := purge(ctx); err != nil {
		return fmt.Errorf("purge save data: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM saves WHERE id = $1`, id); err != nil {
		return fmt.Errorf("delete save: %w", err)
	}

	return tx.Commit(ctx)
}

func (s *Store) DeleteIdleSave(ctx context.Context, id uuid.UUID, idle time.Duration, purge func(context.Context) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin idle save delete: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := lockSave(ctx, tx, id); err != nil {
		return err
	}

	var idleEnough bool
	err = tx.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1 FROM saves s
		WHERE s.id = $1
		  AND COALESCE((SELECT max(m.last_active_at) FROM mc_servers m WHERE m.save_id = s.id), s.created_at) < now() - make_interval(secs => $2)
		  AND NOT EXISTS (SELECT 1 FROM mc_servers m WHERE m.save_id = s.id AND m.save_locked)
	)`, id, idle.Seconds()).Scan(&idleEnough)
	if err != nil {
		return fmt.Errorf("recheck idle save: %w", err)
	}
	if !idleEnough {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM saves WHERE id = $1)`, id).Scan(&exists); err != nil {
			return fmt.Errorf("find idle save: %w", err)
		}
		if !exists {
			return ErrSaveNotFound
		}
		return ErrSaveInUse
	}

	if err := purge(ctx); err != nil {
		return fmt.Errorf("purge save data: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM saves WHERE id = $1`, id); err != nil {
		return fmt.Errorf("delete save: %w", err)
	}

	return tx.Commit(ctx)
}

func (s *Store) ListSavesIdleSince(ctx context.Context, idle time.Duration) ([]uuid.UUID, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT s.id FROM saves s
		 WHERE COALESCE((SELECT max(m.last_active_at) FROM mc_servers m WHERE m.save_id = s.id), s.created_at) < now() - make_interval(secs => $1)
		   AND NOT EXISTS (SELECT 1 FROM mc_servers m WHERE m.save_id = s.id AND m.save_locked)
		 ORDER BY s.created_at, s.id`, idle.Seconds())
	if err != nil {
		return nil, fmt.Errorf("list idle saves: %w", err)
	}
	defer rows.Close()

	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan idle save: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

func (s *Store) LatestSaveDefinition(ctx context.Context, id uuid.UUID) (json.RawMessage, bool, error) {
	var definition json.RawMessage
	err := s.pool.QueryRow(ctx,
		`SELECT definition FROM mc_servers WHERE save_id = $1 AND last_active_at IS NOT NULL ORDER BY created_at DESC, id DESC LIMIT 1`, id,
	).Scan(&definition)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("get latest save definition: %w", err)
	}

	return definition, true, nil
}

func (s *Store) ListReleasableSaveLocks(ctx context.Context) ([]ServerRecord, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name FROM mc_servers WHERE save_locked AND status IN ('stopped', 'expired')`)
	if err != nil {
		return nil, fmt.Errorf("list save locks: %w", err)
	}
	defer rows.Close()

	records := make([]ServerRecord, 0)
	for rows.Next() {
		var record ServerRecord
		if err := rows.Scan(&record.ID, &record.Name); err != nil {
			return nil, fmt.Errorf("scan save lock: %w", err)
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (s *Store) ReleaseSaveLock(ctx context.Context, id uuid.UUID) error {
	if _, err := s.pool.Exec(ctx, `UPDATE mc_servers SET save_locked = false WHERE id = $1 AND status IN ('stopped', 'expired')`, id); err != nil {
		return fmt.Errorf("release save lock: %w", err)
	}

	return nil
}
