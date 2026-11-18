package workspace

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms-app/internal/store"
)

type Postgres struct {
	db *store.DB
}

func NewPostgres(db *store.DB) *Postgres {
	return &Postgres{db: db}
}

func pg(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func pgTime(at time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: at, Valid: true}
}

func recordOf(row store.Workspace) Record {
	return Record{
		ID:       uuid.UUID(row.ID.Bytes),
		Key:      Key{Person: uuid.UUID(row.PersonID.Bytes), Project: uuid.UUID(row.ProjectID.Bytes)},
		Backend:  row.Backend,
		State:    State(row.State),
		Handle:   Handle(row.Handle),
		Reason:   row.Reason,
		LastUsed: row.LastUsedAt.Time,
	}
}

func (p *Postgres) Reserve(ctx context.Context, key Key, backend string, most int, at time.Time) (Record, error) {
	tx, err := p.db.Pool().Begin(ctx)
	if err != nil {
		return Record{}, fmt.Errorf("starting a transaction failed: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queries := p.db.Within(tx).Queries
	if err := queries.LockWorkspaces(ctx); err != nil {
		return Record{}, err
	}
	busy, err := queries.CountBusyWorkspacesBesides(ctx, store.CountBusyWorkspacesBesidesParams{
		PersonID:  pg(key.Person),
		ProjectID: pg(key.Project),
	})
	if err != nil {
		return Record{}, err
	}
	if busy >= int64(most) {
		return Record{}, ErrTooMany
	}

	row, err := queries.StartWorkspace(ctx, store.StartWorkspaceParams{
		PersonID:   pg(key.Person),
		ProjectID:  pg(key.Project),
		Backend:    backend,
		LastUsedAt: pgTime(at),
	})
	if err != nil {
		return Record{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Record{}, fmt.Errorf("finishing the transaction failed: %w", err)
	}
	return recordOf(row), nil
}

func (p *Postgres) Warm(ctx context.Context, id uuid.UUID, handle Handle) error {
	return p.db.Queries.WarmWorkspace(ctx, store.WarmWorkspaceParams{ID: pg(id), Handle: string(handle)})
}

func (p *Postgres) Sleep(ctx context.Context, id uuid.UUID, reason string) error {
	return p.db.Queries.SleepWorkspace(ctx, store.SleepWorkspaceParams{ID: pg(id), Reason: reason})
}

func (p *Postgres) Touch(ctx context.Context, id uuid.UUID, at time.Time) error {
	return p.db.Queries.TouchWorkspace(ctx, store.TouchWorkspaceParams{ID: pg(id), LastUsedAt: pgTime(at)})
}

func (p *Postgres) Busy(ctx context.Context) ([]Record, error) {
	rows, err := p.db.Queries.ListBusyWorkspaces(ctx)
	if err != nil {
		return nil, err
	}
	return recordsOf(rows), nil
}

func (p *Postgres) Sleeping(ctx context.Context) ([]Record, error) {
	rows, err := p.db.Queries.ListSleepingWorkspaces(ctx)
	if err != nil {
		return nil, err
	}
	return recordsOf(rows), nil
}

func (p *Postgres) Remove(ctx context.Context, id uuid.UUID) error {
	return p.db.Queries.DeleteWorkspace(ctx, pg(id))
}

func (p *Postgres) Get(ctx context.Context, key Key) (Record, error) {
	row, err := p.db.Queries.GetWorkspace(ctx, store.GetWorkspaceParams{PersonID: pg(key.Person), ProjectID: pg(key.Project)})
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrNotFound
	}
	if err != nil {
		return Record{}, err
	}
	return recordOf(row), nil
}

func recordsOf(rows []store.Workspace) []Record {
	records := make([]Record, 0, len(rows))
	for _, row := range rows {
		records = append(records, recordOf(row))
	}
	return records
}

var _ Store = (*Postgres)(nil)
