package artifacts

import (
	"context"
	"errors"

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

func id(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: value, Valid: true}
}

func fromRow(row store.Artifact) Artifact {
	artifact := Artifact{
		ID: uuid.UUID(row.ID.Bytes), ProjectID: uuid.UUID(row.ProjectID.Bytes),
		Name: row.Name, Size: row.Size, SHA256: row.Sha256, CreatedAt: row.CreatedAt.Time,
	}
	if row.PublishID.Valid {
		publish := uuid.UUID(row.PublishID.Bytes)
		artifact.PublishID = &publish
	}
	return artifact
}

func fromRows(rows []store.Artifact) []Artifact {
	found := make([]Artifact, 0, len(rows))
	for _, row := range rows {
		found = append(found, fromRow(row))
	}
	return found
}

func (p *Postgres) Insert(ctx context.Context, artifact Artifact) error {
	publish := pgtype.UUID{}
	if artifact.PublishID != nil {
		publish = id(*artifact.PublishID)
	}
	return p.queries.InsertArtifact(ctx, store.InsertArtifactParams{
		ID: id(artifact.ID), ProjectID: id(artifact.ProjectID), PublishID: publish,
		Name: artifact.Name, Size: artifact.Size, Sha256: artifact.SHA256,
	})
}

func (p *Postgres) List(ctx context.Context, projectID uuid.UUID, limit int32) ([]Artifact, error) {
	rows, err := p.queries.ListArtifacts(ctx, store.ListArtifactsParams{ProjectID: id(projectID), Limit: limit})
	return fromRows(rows), err
}

func (p *Postgres) Get(ctx context.Context, projectID, artifactID uuid.UUID) (Artifact, error) {
	row, err := p.queries.GetArtifact(ctx, store.GetArtifactParams{ProjectID: id(projectID), ID: id(artifactID)})
	if errors.Is(err, pgx.ErrNoRows) {
		return Artifact{}, ErrMissing
	}
	if err != nil {
		return Artifact{}, err
	}
	return fromRow(row), nil
}

func (p *Postgres) OfPublish(ctx context.Context, publishID uuid.UUID) ([]Artifact, error) {
	rows, err := p.queries.ArtifactsOfPublish(ctx, id(publishID))
	return fromRows(rows), err
}

func (p *Postgres) Any(ctx context.Context, projectID uuid.UUID) (bool, error) {
	return p.queries.ProjectHasArtifacts(ctx, id(projectID))
}

var _ Records = (*Postgres)(nil)
