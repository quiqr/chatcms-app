package preview

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/policy"
	"github.com/quiqr/chatcms-app/internal/tenancy"
)

// Web answers the HTTP layer's questions.
type Web struct {
	service *Service
	tenancy *tenancy.Store
}

func NewWeb(service *Service, tenancyStore *tenancy.Store) (*Web, error) {
	if service == nil || tenancyStore == nil {
		return nil, ErrIncomplete
	}
	return &Web{service: service, tenancy: tenancyStore}, nil
}

func (w *Web) PreviewHandler(ctx context.Context, who httpapi.Actor, organisation, project, under string) (http.Handler, error) {
	if who.Anonymous || who.ID == "" {
		return nil, httpapi.ErrNotAllowed
	}
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return nil, httpapi.ErrNotAllowed
	}

	actor, err := w.tenancy.ActorFor(ctx, personID)
	if err != nil {
		return nil, err
	}

	found, err := w.tenancy.ProjectBySlug(ctx, organisation, project)
	if errors.Is(err, tenancy.ErrNotFound) {
		return nil, httpapi.ErrMissing
	}
	if err != nil {
		return nil, err
	}

	return w.service.Handler(ctx, actor, found.ID, under)
}

func (w *Web) project(ctx context.Context, who httpapi.Actor, organisation, project string) (policy.Actor, uuid.UUID, error) {
	if who.Anonymous || who.ID == "" {
		return policy.Actor{}, uuid.UUID{}, httpapi.ErrNotAllowed
	}
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return policy.Actor{}, uuid.UUID{}, httpapi.ErrNotAllowed
	}
	actor, err := w.tenancy.ActorFor(ctx, personID)
	if err != nil {
		return policy.Actor{}, uuid.UUID{}, err
	}
	found, err := w.tenancy.ProjectBySlug(ctx, organisation, project)
	if errors.Is(err, tenancy.ErrNotFound) {
		return policy.Actor{}, uuid.UUID{}, httpapi.ErrMissing
	}
	if err != nil {
		return policy.Actor{}, uuid.UUID{}, err
	}
	return actor, found.ID, nil
}

func (w *Web) OpenWorkspace(ctx context.Context, who httpapi.Actor, organisation, project string) (httpapi.WorkspaceView, error) {
	actor, projectID, err := w.project(ctx, who, organisation, project)
	if err != nil {
		return httpapi.WorkspaceView{}, err
	}
	standing, err := w.service.Open(ctx, actor, projectID, httpapi.PreviewPrefix+organisation+"/"+project+"/")
	if err != nil {
		return httpapi.WorkspaceView{}, err
	}
	return viewOf(standing), nil
}

func (w *Web) WorkspaceStanding(ctx context.Context, who httpapi.Actor, organisation, project string) (httpapi.WorkspaceView, error) {
	actor, projectID, err := w.project(ctx, who, organisation, project)
	if err != nil {
		return httpapi.WorkspaceView{}, err
	}
	standing, err := w.service.Standing(ctx, actor, projectID)
	if err != nil {
		return httpapi.WorkspaceView{}, err
	}
	return viewOf(standing), nil
}

func viewOf(standing Standing) httpapi.WorkspaceView {
	view := httpapi.WorkspaceView{State: string(standing.State), Reason: standing.Reason, Log: standing.Log, Address: standing.Address, Type: standing.Type, Open: standing.Open, OwnPages: standing.OwnPages}
	for _, step := range standing.Steps {
		view.Steps = append(view.Steps, httpapi.WorkspaceStepView{Step: string(step.Step), State: string(step.State)})
	}
	return view
}

var _ httpapi.PreviewService = (*Web)(nil)
var _ httpapi.WorkspaceService = (*Web)(nil)
