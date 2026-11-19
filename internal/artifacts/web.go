package artifacts

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/policy"
	"github.com/quiqr/chatcms-app/internal/tenancy"
)

type Web struct {
	store   *Store
	tenancy *tenancy.Store
}

func NewWeb(store *Store, tenancyStore *tenancy.Store) *Web {
	return &Web{store: store, tenancy: tenancyStore}
}

func (w *Web) project(ctx context.Context, who httpapi.Actor, organisation, project string) (uuid.UUID, error) {
	if who.Anonymous || who.ID == "" {
		return uuid.Nil, httpapi.ErrNotAllowed
	}
	person, err := uuid.Parse(who.ID)
	if err != nil {
		return uuid.Nil, httpapi.ErrNotAllowed
	}
	actor, err := w.tenancy.ActorFor(ctx, person)
	if err != nil {
		return uuid.Nil, err
	}
	found, err := w.tenancy.ProjectBySlug(ctx, organisation, project)
	if errors.Is(err, tenancy.ErrNotFound) {
		return uuid.Nil, httpapi.ErrMissing
	}
	if err != nil {
		return uuid.Nil, err
	}
	if !policy.Can(actor, policy.ActionReadContent, policy.Project(found.OrganisationID.String(), found.ID.String())) {
		return uuid.Nil, httpapi.ErrMissing
	}
	return found.ID, nil
}

func View(artifact Artifact) httpapi.ArtifactView {
	view := httpapi.ArtifactView{ID: artifact.ID.String(), Name: artifact.Name, Size: artifact.Size, CreatedAt: artifact.CreatedAt.UTC().Format(time.RFC3339)}
	if artifact.PublishID != nil {
		view.PublishID = artifact.PublishID.String()
	}
	return view
}

func (w *Web) Artifacts(ctx context.Context, who httpapi.Actor, organisation, project string) ([]httpapi.ArtifactView, error) {
	projectID, err := w.project(ctx, who, organisation, project)
	if err != nil {
		return nil, err
	}
	list, err := w.store.List(ctx, projectID)
	if err != nil {
		return nil, err
	}
	views := make([]httpapi.ArtifactView, 0, len(list))
	for _, artifact := range list {
		views = append(views, View(artifact))
	}
	return views, nil
}

func (w *Web) OpenArtifact(ctx context.Context, who httpapi.Actor, organisation, project, id string) (httpapi.ArtifactView, io.ReadSeekCloser, time.Time, error) {
	projectID, err := w.project(ctx, who, organisation, project)
	if err != nil {
		return httpapi.ArtifactView{}, nil, time.Time{}, err
	}
	artifactID, err := uuid.Parse(id)
	if err != nil {
		return httpapi.ArtifactView{}, nil, time.Time{}, httpapi.ErrMissing
	}
	artifact, file, err := w.store.Open(ctx, projectID, artifactID)
	if errors.Is(err, ErrMissing) {
		return httpapi.ArtifactView{}, nil, time.Time{}, httpapi.ErrMissing
	}
	if err != nil {
		return httpapi.ArtifactView{}, nil, time.Time{}, err
	}
	return View(artifact), file, artifact.CreatedAt, nil
}

var _ httpapi.ArtifactService = (*Web)(nil)
