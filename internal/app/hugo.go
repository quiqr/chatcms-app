package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/nixpin"
	"github.com/quiqr/chatcms-app/internal/runtime"
	"github.com/quiqr/chatcms-app/internal/store"
)

type hugos struct {
	installed string
	versionOf func(ctx context.Context, projectID uuid.UUID) (string, error)
	provided  map[string]string
	pins      packagePins
}

type packagePins interface {
	Path(ctx context.Context, name, version, revision string) (string, error)
	Environment(ctx context.Context, kind, revision string, packages []string) (string, error)
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
	path, err := h.pins.Path(ctx, "hugo", version, "")
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

type namedPins struct {
	pins packagePins
}

func (n namedPins) Environment(ctx context.Context, kind, revision string, packages []string) (string, error) {
	path, err := n.pins.Environment(ctx, kind, revision, packages)
	if err == nil {
		return path, nil
	}
	detail := map[string]any{"package": kind, "version": strings.Join(packages, ", ")}
	if errors.Is(err, nixpin.ErrNotFound) || errors.Is(err, nixpin.ErrUnusable) {
		return "", &httpapi.RefusedError{Key: "refused.package-not-found", Reason: fmt.Sprintf("the %s packages %s could not be found", kind, strings.Join(packages, ", ")), Detail: detail}
	}
	return "", &httpapi.RefusedError{Key: "refused.package-unavailable", Reason: fmt.Sprintf("the %s packages could not be made ready; try again in a few minutes", kind), Detail: detail}
}

func (n namedPins) Path(ctx context.Context, name, version, revision string) (string, error) {
	path, err := n.pins.Path(ctx, name, version, revision)
	if err == nil {
		return path, nil
	}
	notFound, unavailable := "refused.package-not-found", "refused.package-unavailable"
	if name == "hugo" {
		notFound, unavailable = "refused.hugo-version-not-found", "refused.hugo-version-unavailable"
	}
	detail := map[string]any{"version": version, "package": name}
	if errors.Is(err, nixpin.ErrNotFound) || errors.Is(err, nixpin.ErrUnusable) {
		return "", &httpapi.RefusedError{Key: notFound, Reason: fmt.Sprintf("%s %s could not be found", name, version), Detail: detail}
	}
	return "", &httpapi.RefusedError{Key: unavailable, Reason: fmt.Sprintf("%s %s could not be made ready; try again in a few minutes", name, version), Detail: detail}
}

func (h hugos) Runtime(ctx context.Context, projectID uuid.UUID, site fs.FS) (runtime.Resolved, error) {
	version, err := h.versionOf(ctx, projectID)
	if err != nil {
		return runtime.Resolved{}, err
	}
	definition, declared, err := runtime.Read(site, version)
	if err != nil {
		return runtime.Resolved{}, &httpapi.RefusedError{Key: "refused.runtime-definition", Reason: err.Error(), Detail: map[string]any{"reason": err.Error()}}
	}
	resolver := runtime.Resolver{InstalledHugo: h.installed, Provided: h.provided}
	if h.pins != nil {
		resolver.Pins = namedPins{pins: h.pins}
	}
	resolved, err := resolver.Resolve(ctx, definition, declared)
	switch {
	case errors.Is(err, runtime.ErrNoProgram), errors.Is(err, runtime.ErrNoPackages), errors.Is(err, runtime.ErrNoHugo):
		return runtime.Resolved{}, &httpapi.RefusedError{Key: "refused.runtime-definition", Reason: err.Error(), Detail: map[string]any{"reason": err.Error()}}
	case err != nil:
		return runtime.Resolved{}, err
	}
	return resolved, nil
}
