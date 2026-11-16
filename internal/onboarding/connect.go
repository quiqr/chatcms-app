package onboarding

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/store"
)

var (
	ErrNotKept               = errors.New("this site is already connected to a repository")
	ErrConnectNeedsAnAddress = errors.New("connecting needs the address of an empty repository")
	ErrConnectNotEmpty       = errors.New("that repository is not empty; connecting needs an empty one")
	ErrConnectKeyRefused     = errors.New("that repository did not accept the project's key")
	ErrConnectUnreachable    = errors.New("that repository could not be reached")
)

func (i *Importer) Connect(ctx context.Context, projectID uuid.UUID, remote string) error {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return ErrConnectNeedsAnAddress
	}

	row, err := i.db.Queries.GetProject(ctx, pgtype.UUID{Bytes: projectID, Valid: true})
	if err != nil {
		return fmt.Errorf("reading the project failed: %w", err)
	}
	if !row.Hosted {
		return ErrNotKept
	}

	siteID := projectID.String()
	if err := reached(i.sites.Reachable(ctx, siteID, remote)); err != nil {
		return err
	}
	empty, err := i.sites.RemoteIsEmpty(ctx, siteID, remote)
	if err != nil {
		return reached(err)
	}
	if !empty {
		return ErrConnectNotEmpty
	}

	repository, err := i.sites.Open(ctx, siteID)
	if err != nil {
		return err
	}
	defer repository.Close()

	if err := repository.MoveTo(ctx, remote); err != nil {
		return fmt.Errorf("%w: %v", ErrConnectUnreachable, err)
	}

	if err := i.db.Queries.ConnectProject(ctx, store.ConnectProjectParams{
		ID:     pgtype.UUID{Bytes: projectID, Valid: true},
		Remote: remote,
	}); err != nil {
		return fmt.Errorf("recording the repository failed: %w", err)
	}
	return i.sites.Unhost(siteID)
}

func reached(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gitsync.ErrKeyRefused):
		return fmt.Errorf("%w: %v", ErrConnectKeyRefused, err)
	default:
		return fmt.Errorf("%w: %v", ErrConnectUnreachable, err)
	}
}
