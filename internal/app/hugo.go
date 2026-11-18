package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/nixpin"
	"github.com/quiqr/chatcms-app/internal/store"
)

type hugos struct {
	installed string
	versionOf func(ctx context.Context, projectID uuid.UUID) (string, error)
	pins      interface {
		Path(ctx context.Context, name, version string) (string, error)
	}
}

func projectHugoVersion(queries *store.Queries) func(ctx context.Context, projectID uuid.UUID) (string, error) {
	return func(ctx context.Context, projectID uuid.UUID) (string, error) {
		row, err := queries.GetProject(ctx, pgUUID(projectID))
		return row.HugoVersion, err
	}
}

func (h hugos) For(ctx context.Context, projectID uuid.UUID) (string, error) {
	version, err := h.versionOf(ctx, projectID)
	if err != nil {
		return "", err
	}
	if version == "" || h.pins == nil {
		return h.installed, nil
	}
	path, err := h.pins.Path(ctx, "hugo", version)
	switch {
	case errors.Is(err, nixpin.ErrNotFound), errors.Is(err, nixpin.ErrUnusable):
		return "", &httpapi.RefusedError{
			Key:    "refused.hugo-version-not-found",
			Reason: fmt.Sprintf("Hugo %s could not be found; choose another version in the project settings", version),
			Detail: map[string]any{"version": version},
		}
	case err != nil:
		return "", &httpapi.RefusedError{
			Key:    "refused.hugo-version-unavailable",
			Reason: fmt.Sprintf("Hugo %s could not be made ready; try again in a few minutes", version),
			Detail: map[string]any{"version": version},
		}
	}
	return filepath.Join(path, "bin", "hugo"), nil
}
