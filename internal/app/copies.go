package app

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/jobs"
	"github.com/quiqr/chatcms-app/internal/sitesync"
)

const ShareJobKind = "site-share"

var errAnswerFirst = &httpapi.RefusedError{
	Key:    "refused.answer-the-question-first",
	Reason: "a page was edited by two people; answer the question about it first",
}

func openFor(ctx context.Context, sites *gitsync.Sites, personID string, projectID uuid.UUID) (gitsync.Repo, error) {
	repository, err := sites.OpenFor(ctx, projectID.String(), personID)
	if errors.Is(err, gitsync.ErrNotAWorkingCopy) {
		return nil, errNothingBroughtIn
	}
	if errors.Is(err, gitsync.ErrUnusableSiteID) {
		return nil, httpapi.ErrNotAllowed
	}
	return repository, err
}

func (p publishSources) openFor(ctx context.Context, personID string, projectID uuid.UUID) (gitsync.Repo, error) {
	return openFor(ctx, p.sites, personID, projectID)
}

func (p publishSources) Share(ctx context.Context, personID string, projectID uuid.UUID) error {
	result, err := p.sites.Share(ctx, projectID.String(), personID)
	if errors.Is(err, gitsync.ErrExchangeInProgress) || (err == nil && result.Conflict != nil) {
		return errAnswerFirst
	}
	if errors.Is(err, gitsync.ErrNotAWorkingCopy) {
		return errNothingBroughtIn
	}
	return err
}

func shareSubject(personID string, projectID uuid.UUID) string {
	return projectID.String() + "/" + personID
}

func parseShareSubject(subject string) (string, string, bool) {
	project, person, found := strings.Cut(subject, "/")
	return project, person, found && project != "" && person != ""
}

func enqueueShare(ctx context.Context, queue *jobs.Queue, personID string, projectID uuid.UUID) error {
	if queue == nil {
		return nil
	}
	_, err := queue.Enqueue(ctx, ShareJobKind, shareSubject(personID, projectID), nil)
	return err
}

type sharedAfterGoingBack struct {
	historyCopy
	queue    *jobs.Queue
	personID string
	project  uuid.UUID
}

func (s sharedAfterGoingBack) GoBackTo(ctx context.Context, version string, author gitsync.Identity, message string) (gitsync.Version, error) {
	made, err := s.historyCopy.GoBackTo(ctx, version, author, message)
	if err == nil {
		_ = enqueueShare(ctx, s.queue, s.personID, s.project)
	}
	return made, err
}

type standingCopy struct {
	workingCopy
	shared   gitsync.Repo
	sites    *gitsync.Sites
	site     string
	personID string
}

func (c standingCopy) Status(ctx context.Context) (gitsync.Status, error) {
	mine, err := c.Repo.Status(ctx)
	if err != nil {
		return gitsync.Status{}, err
	}
	theirs, err := c.shared.Status(ctx)
	if err != nil {
		return gitsync.Status{}, err
	}
	return gitsync.Status{
		UnsavedWork:        mine.UnsavedWork,
		Unpublished:        mine.Unpublished,
		RemoteAhead:        theirs.RemoteAhead,
		ExchangeInProgress: mine.ExchangeInProgress || theirs.ExchangeInProgress,
	}, nil
}

func (c standingCopy) sharedAsking(ctx context.Context) bool {
	status, err := c.shared.Status(ctx)
	return err == nil && status.ExchangeInProgress
}

func (c standingCopy) Sync(ctx context.Context) (gitsync.SyncResult, error) {
	if !c.sharedAsking(ctx) {
		if mine, err := c.Repo.Status(ctx); err == nil && mine.ExchangeInProgress {
			return c.Repo.Sync(ctx)
		}
	}
	result, err := c.shared.Sync(ctx)
	if errors.Is(err, gitsync.ErrRemoteUnreachable) {
		shared, shareErr := c.sites.Share(ctx, c.site, c.personID)
		if shareErr != nil {
			return shared, shareErr
		}
		return shared, err
	}
	if err != nil || result.Conflict != nil {
		return result, err
	}
	return c.sites.Share(ctx, c.site, c.personID)
}

func (c standingCopy) Resolve(ctx context.Context, resolutions map[string][]byte) (gitsync.SyncResult, error) {
	if c.sharedAsking(ctx) {
		return c.shared.Resolve(ctx, resolutions)
	}
	result, err := c.Repo.Resolve(ctx, resolutions)
	if err != nil || result.Conflict != nil {
		return result, err
	}
	return c.sites.Share(ctx, c.site, c.personID)
}

func (c standingCopy) Abandon(ctx context.Context) error {
	if c.sharedAsking(ctx) {
		return c.shared.Abandon(ctx)
	}
	return c.Repo.Abandon(ctx)
}

var _ sitesync.Site = standingCopy{}
