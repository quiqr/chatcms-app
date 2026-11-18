package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/policy"
	"github.com/quiqr/chatcms-app/internal/tenancy"
	"github.com/quiqr/chatcms-app/internal/workspace"
	"github.com/quiqr/chatcms-app/internal/workspace/local"
)

type workspaceCopies struct {
	sites *gitsync.Sites
}

func (c workspaceCopies) Dir(ctx context.Context, key workspace.Key) (string, error) {
	repository, err := openFor(ctx, c.sites, key.Person.String(), key.Project)
	if err != nil {
		return "", err
	}
	repository.Close()
	return c.sites.DirFor(key.Project.String(), key.Person.String())
}

func (c workspaceCopies) Unshared(ctx context.Context, key workspace.Key) (int, error) {
	return c.sites.Unshared(ctx, key.Project.String(), key.Person.String())
}

func (c workspaceCopies) Remove(ctx context.Context, key workspace.Key) error {
	return c.sites.RemoveFor(key.Project.String(), key.Person.String())
}

var _ workspace.Copies = workspaceCopies{}

func mayStillUse(sources publishSources, people *tenancy.Store) workspace.Allowed {
	return func(ctx context.Context, key workspace.Key) (bool, error) {
		row, err := sources.queries.GetProject(ctx, pgUUID(key.Project))
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if row.ArchivedAt.Valid {
			return false, nil
		}
		actor, err := people.ActorFor(ctx, key.Person)
		if errors.Is(err, tenancy.ErrNotFound) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		resource := policy.Project(uuid.UUID(row.OrganisationID.Bytes).String(), key.Project.String())
		return policy.Can(actor, policy.ActionReadContent, resource), nil
	}
}

func (s *Server) startWorkspaces(ctx context.Context, tenancyStore *tenancy.Store) (*workspace.Workspaces, error) {
	backend, err := local.New(local.Settings{HugoBin: s.cfg.HugoBin})
	if err != nil {
		return nil, err
	}
	sources := publishSources{sites: s.sites, queries: s.db.Queries, baseURL: s.cfg.BaseURL}
	made, err := workspace.New(
		backend,
		workspace.NewPostgres(s.db),
		workspaceCopies{sites: s.sites},
		mayStillUse(sources, tenancyStore),
		workspace.Settings{Most: s.cfg.WorkspacesAtOnce, Idle: s.cfg.WorkspaceIdle, Logger: s.logger},
	)
	if err != nil {
		return nil, err
	}
	if err := made.Reconcile(ctx); err != nil {
		return nil, err
	}
	made.Run()
	return made, nil
}
