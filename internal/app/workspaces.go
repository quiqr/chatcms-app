package app

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/nixpin"
	"github.com/quiqr/chatcms-app/internal/policy"
	"github.com/quiqr/chatcms-app/internal/runtime"
	"github.com/quiqr/chatcms-app/internal/tenancy"
	"github.com/quiqr/chatcms-app/internal/workspace"
	"github.com/quiqr/chatcms-app/internal/workspace/gvisor"
	"github.com/quiqr/chatcms-app/internal/workspace/local"
	"github.com/quiqr/chatcms-app/internal/workspace/microvm"
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

func (s *Server) workspaceBackend() (workspace.Backend, error) {
	if s.cfg.WorkspaceBackend == "gvisor" {
		gateway := ""
		if s.aiListener != nil {
			gateway = s.aiListener.Addr().String()
		}
		return gvisor.New(gvisor.Settings{
			Runsc:    s.cfg.Runsc,
			Relay:    s.cfg.RelayBin,
			OpenCode: s.cfg.OpenCodeBin,
			HugoBin:  s.cfg.HugoBin,
			State:    s.cfg.WorkspaceState,
			Gateway:  gateway,
		})
	}
	if s.cfg.WorkspaceBackend != "microvm" {
		gateway := ""
		if s.aiListener != nil {
			gateway = "http://" + s.aiListener.Addr().String()
		}
		return local.New(local.Settings{HugoBin: s.cfg.HugoBin, OpenCodeBin: s.cfg.OpenCodeBin, Gateway: gateway})
	}
	guest, err := microvm.LoadGuest(s.cfg.WorkspaceGuest)
	if err != nil {
		return nil, err
	}
	return microvm.New(microvm.Settings{
		CloudHypervisor: s.cfg.CloudHypervisor,
		Virtiofsd:       s.cfg.VirtiofsdBin,
		Guest:           guest,
		RunDir:          s.cfg.WorkspaceRunDir,
		Taps:            s.cfg.WorkspaceTaps,
		HugoBin:         s.cfg.HugoBin,
		GatewayPort:     s.gatewayPort(),
	})
}

func (s *Server) projectHugo() hugos {
	if s.hugo != nil {
		return *s.hugo
	}
	installed := s.cfg.HugoBin
	if found, err := exec.LookPath(installed); err == nil {
		installed = found
	}
	if absolute, err := filepath.Abs(installed); err == nil {
		installed = absolute
	}
	found := hugos{installed: installed, versionOf: projectHugoVersion(s.db.Queries), provided: s.cfg.Packages}
	s.hugo = &found
	if s.cfg.NixBin == "" {
		return found
	}
	pins, err := nixpin.New(nixpin.NewPostgres(s.db.Queries), nixpin.Settings{
		Nixhub: s.cfg.NixhubURL,
		NixBin: s.cfg.NixBin,
		Roots:  filepath.Join(s.cfg.StateDir, "gcroots"),
	})
	if err != nil {
		s.logger.Warn("a project's own Hugo version cannot be used", "reason", err)
		return found
	}
	found.pins = pins
	s.hugo = &found
	return found
}

func (s *Server) startWorkspaces(ctx context.Context, tenancyStore *tenancy.Store) (*workspace.Workspaces, error) {
	backend, err := s.workspaceBackend()
	if err != nil {
		return nil, err
	}
	hugo := s.projectHugo()
	sources := publishSources{sites: s.sites, queries: s.db.Queries, baseURL: s.cfg.BaseURL}
	history := workspace.NewPostgres(s.db)
	made, err := workspace.New(
		backend,
		history,
		workspaceCopies{sites: s.sites},
		mayStillUse(sources, tenancyStore),
		workspace.Settings{
			Most:   s.cfg.WorkspacesAtOnce,
			Idle:   s.cfg.WorkspaceIdle,
			Logger: s.logger,
			RuntimeFor: func(ctx context.Context, key workspace.Key, dir string) (runtime.Resolved, error) {
				return hugo.Runtime(ctx, key.Project, os.DirFS(dir))
			},
			Tokens:      s.workspaceTokens(),
			History:     history,
			HistoryDays: s.cfg.HistoryDays,
		},
	)
	if err != nil {
		return nil, err
	}
	if err := history.Interrupt(ctx); err != nil {
		return nil, err
	}
	if err := made.Reconcile(ctx); err != nil {
		return nil, err
	}
	made.Run()
	return made, nil
}

type workspaceStatistics struct {
	store *workspace.Postgres
	sites string
	limit int
	days  int
	clock func() time.Time
}

const statisticsDays = 7

func (w workspaceStatistics) WorkspaceStatistics(ctx context.Context) (any, error) {
	stats, err := w.store.Statistics(ctx, w.clock().AddDate(0, 0, -statisticsDays))
	if err != nil {
		return nil, err
	}
	stats.Limit = w.limit
	stats.Days = statisticsDays
	stats.HistoryDays = w.days
	for i := range stats.Projects {
		stats.Projects[i].DiskBytes, stats.Projects[i].DiskKnown = workspace.DiskOf(filepath.Join(w.sites, stats.Projects[i].ID.String(), "people"), 5*time.Second)
	}
	return stats, nil
}
