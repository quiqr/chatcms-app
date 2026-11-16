package onboarding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/quiqr/chatcms-app/internal/store"
)

const (
	TemplateRefreshJobKind = "templates-refresh"
	DefaultTemplateFeed    = "https://quiqr.github.io/quiqr-community-templates/templates.json"

	refreshAfter       = time.Hour
	retryAfter         = 5 * time.Minute
	feedSizeLimit      = 4 << 20
	screenshotLimit    = 8 << 20
	screenshotDirName  = "template-screenshots"
	fetchTimeout       = 20 * time.Second
	refreshJobSubject  = "feed"
	screenshotFileMode = 0o600
)

var (
	ErrGalleryOff         = errors.New("this installation offers no templates")
	ErrGalleryUnavailable = errors.New("the template list could not be loaded")
	ErrNoScreenshot       = errors.New("this template has no screenshot")
)

type Scheduler interface {
	EnqueueAt(ctx context.Context, kind, subject string, payload any, when time.Time) (int64, error)
}

type Catalogue struct {
	db       *store.DB
	feedURL  string
	shotsDir string
	client   *http.Client
	now      func() time.Time
}

func NewCatalogue(db *store.DB, feedURL, stateDir string, client *http.Client) *Catalogue {
	if client == nil {
		client = &http.Client{Timeout: fetchTimeout}
	}
	return &Catalogue{
		db:       db,
		feedURL:  strings.TrimSpace(feedURL),
		shotsDir: filepath.Join(stateDir, screenshotDirName),
		client:   client,
		now:      time.Now,
	}
}

func (c *Catalogue) Enabled() bool {
	return c != nil && c.feedURL != ""
}

func (c *Catalogue) Refresh(ctx context.Context) error {
	if !c.Enabled() {
		return ErrGalleryOff
	}

	data, err := c.fetch(ctx, c.feedURL, feedSizeLimit)
	if err != nil {
		return err
	}
	templates, err := ParseFeed(data, c.feedURL)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(templates)
	if err != nil {
		return err
	}
	return c.db.Queries.SaveTemplateFeed(ctx, store.SaveTemplateFeedParams{
		FeedUrl:   c.feedURL,
		Templates: encoded,
	})
}

func (c *Catalogue) RefreshJob(scheduler Scheduler, report func(error)) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		err := c.Refresh(ctx)
		if err != nil && report != nil {
			report(err)
		}

		next := refreshAfter
		if err != nil {
			if _, haveCopy := c.stored(ctx); !haveCopy {
				next = retryAfter
			}
		}
		_, scheduleErr := scheduler.EnqueueAt(ctx, TemplateRefreshJobKind, refreshJobSubject, nil, c.now().Add(next))
		return scheduleErr
	}
}

func (c *Catalogue) Start(ctx context.Context, scheduler Scheduler) error {
	if !c.Enabled() {
		return nil
	}
	_, err := scheduler.EnqueueAt(ctx, TemplateRefreshJobKind, refreshJobSubject, nil, c.now())
	return err
}

func (c *Catalogue) List(ctx context.Context) ([]Template, error) {
	if !c.Enabled() {
		return nil, ErrGalleryOff
	}
	feed, ok := c.stored(ctx)
	if !ok {
		return nil, ErrGalleryUnavailable
	}
	return feed.templates, nil
}

func (c *Catalogue) Find(ctx context.Context, id string) (Template, error) {
	templates, err := c.List(ctx)
	if err != nil {
		return Template{}, err
	}
	for _, template := range templates {
		if template.ID == id {
			return template, nil
		}
	}
	return Template{}, fmt.Errorf("%w: %q", ErrNoSuchTemplate, id)
}

type storedFeed struct {
	templates []Template
	fetchedAt time.Time
}

func (c *Catalogue) stored(ctx context.Context) (storedFeed, bool) {
	row, err := c.db.Queries.GetTemplateFeed(ctx)
	if err != nil || row.FeedUrl != c.feedURL {
		return storedFeed{}, false
	}
	var templates []Template
	if json.Unmarshal(row.Templates, &templates) != nil {
		return storedFeed{}, false
	}
	return storedFeed{templates: templates, fetchedAt: row.FetchedAt.Time}, true
}

func (c *Catalogue) Screenshot(ctx context.Context, id string) ([]byte, string, error) {
	if !c.Enabled() {
		return nil, "", ErrGalleryOff
	}
	feed, ok := c.stored(ctx)
	if !ok {
		return nil, "", ErrGalleryUnavailable
	}

	var template Template
	for _, candidate := range feed.templates {
		if candidate.ID == id {
			template = candidate
		}
	}
	if template.ID == "" {
		return nil, "", fmt.Errorf("%w: %q", ErrNoSuchTemplate, id)
	}

	path := filepath.Join(c.shotsDir, fmt.Sprintf("%s@%d", template.ID, feed.fetchedAt.UnixMicro()))
	if data, err := os.ReadFile(path); err == nil {
		if kind, image := imageKind(data); image {
			return data, kind, nil
		}
	}

	address, ok := ScreenshotURL(c.feedURL, template)
	if !ok {
		return nil, "", ErrNoScreenshot
	}
	data, err := c.fetch(ctx, address, screenshotLimit)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrNoScreenshot, err)
	}
	kind, image := imageKind(data)
	if !image {
		return nil, "", fmt.Errorf("%w: what came back is not a picture", ErrNoScreenshot)
	}

	if err := os.MkdirAll(c.shotsDir, 0o700); err == nil {
		older, _ := filepath.Glob(filepath.Join(c.shotsDir, template.ID+"@*"))
		for _, stale := range older {
			_ = os.Remove(stale)
		}
		_ = os.WriteFile(path, data, screenshotFileMode)
	}
	return data, kind, nil
}

func imageKind(data []byte) (string, bool) {
	kind := http.DetectContentType(data)
	switch kind {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
		return kind, true
	}
	return "", false
}

func (c *Catalogue) fetch(ctx context.Context, address string, limit int64) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetching %s failed: %w", address, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %s answered %s", address, response.Status)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("reading %s failed: %w", address, err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%s is larger than %d bytes", address, limit)
	}
	return data, nil
}
