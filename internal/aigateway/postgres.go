package aigateway

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms-app/internal/store"
)

type Postgres struct {
	queries *store.Queries
}

func NewPostgres(queries *store.Queries) *Postgres {
	return &Postgres{queries: queries}
}

func pg(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: id != uuid.Nil}
}

func (p *Postgres) Issue(ctx context.Context, person, project uuid.UUID, hash []byte) error {
	return p.queries.IssueAIToken(ctx, store.IssueAITokenParams{TokenHash: hash, PersonID: pg(person), ProjectID: pg(project)})
}

func (p *Postgres) Revoke(ctx context.Context, person, project uuid.UUID) error {
	return p.queries.RevokeAITokensFor(ctx, store.RevokeAITokensForParams{PersonID: pg(person), ProjectID: pg(project)})
}

func (p *Postgres) LookUp(ctx context.Context, hash []byte) (Caller, bool, error) {
	row, err := p.queries.LookUpAIToken(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return Caller{}, false, nil
	}
	if err != nil {
		return Caller{}, false, err
	}
	return Caller{
		Person:       uuid.UUID(row.PersonID.Bytes),
		Project:      uuid.UUID(row.ProjectID.Bytes),
		Organisation: uuid.UUID(row.OrganisationID.Bytes),
	}, true, nil
}

func (p *Postgres) Record(ctx context.Context, usage Usage) error {
	return p.queries.RecordAIUsage(ctx, store.RecordAIUsageParams{
		OrganisationID: pg(usage.Caller.Organisation),
		ProjectID:      pg(usage.Caller.Project),
		PersonID:       pg(usage.Caller.Person),
		Model:          usage.Model,
		TokensIn:       usage.TokensIn,
		TokensOut:      usage.TokensOut,
		Status:         int32(usage.Status),
	})
}

func (p *Postgres) UsedSince(ctx context.Context, organisation uuid.UUID, since time.Time) (int64, int64, error) {
	row, err := p.queries.AIUsageSince(ctx, store.AIUsageSinceParams{
		OrganisationID: pg(organisation),
		At:             pgtype.Timestamptz{Time: since, Valid: true},
	})
	return row.Requests, row.Tokens, err
}

func (p *Postgres) Limits(ctx context.Context, organisation uuid.UUID) (Limits, error) {
	row, err := p.queries.AILimitsOf(ctx, pg(organisation))
	if errors.Is(err, pgx.ErrNoRows) {
		return Limits{}, ErrNotFound
	}
	if err != nil {
		return Limits{}, err
	}
	limits := Limits{TokensPerMonth: row.AiTokensPerMonth}
	if row.AiRequestsPerMinute != nil {
		requests := int(*row.AiRequestsPerMinute)
		limits.RequestsPerMinute = &requests
	}
	return limits, nil
}

func (p *Postgres) SetLimits(ctx context.Context, organisation uuid.UUID, limits Limits) error {
	var requests *int32
	if limits.RequestsPerMinute != nil {
		value := int32(*limits.RequestsPerMinute)
		requests = &value
	}
	changed, err := p.queries.SetAILimits(ctx, store.SetAILimitsParams{
		ID:                  pg(organisation),
		AiRequestsPerMinute: requests,
		AiTokensPerMonth:    limits.TokensPerMonth,
	})
	if err != nil {
		return err
	}
	if changed == 0 {
		return ErrNotFound
	}
	return nil
}

var _ Store = (*Postgres)(nil)
