package workspace

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms-app/internal/store"
)

func optionalID(id uuid.UUID) pgtype.UUID {
	if id == uuid.Nil {
		return pgtype.UUID{}
	}
	return pg(id)
}

func (p *Postgres) Begin(ctx context.Context, event Event) (int64, error) {
	return p.db.Queries.BeginWorkspaceStart(ctx, store.BeginWorkspaceStartParams{
		At:        pgTime(event.At),
		Backend:   event.Backend,
		ProjectID: optionalID(event.Project),
		PersonID:  optionalID(event.Person),
	})
}

func (p *Postgres) Finish(ctx context.Context, id int64, event Event) error {
	steps, err := json.Marshal(event.Steps)
	if err != nil {
		return err
	}
	var hugo []byte
	if event.Hugo != nil {
		if hugo, err = json.Marshal(event.Hugo); err != nil {
			return err
		}
	}
	millis := int32(event.Millis)
	return p.db.Queries.FinishWorkspaceStart(ctx, store.FinishWorkspaceStartParams{
		ID:         id,
		Kind:       string(event.Kind),
		Millis:     &millis,
		Reason:     event.Reason,
		Steps:      steps,
		Log:        strings.Join(event.Log, "\n"),
		Hugo:       hugo,
		FinishedAt: pgTime(event.At.Add(millisecondsOf(event.Millis))),
	})
}

func (p *Postgres) Interrupt(ctx context.Context) error {
	return p.db.Queries.InterruptWorkspaceStarts(ctx)
}

func millisecondsOf(millis int) time.Duration {
	return time.Duration(millis) * time.Millisecond
}
