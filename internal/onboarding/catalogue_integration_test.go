//go:build integration

package onboarding_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/quiqr/chatcms-app/internal/onboarding"
)

var aPNG = append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 64)...)

type feedServer struct {
	*httptest.Server
	failing     atomic.Bool
	shotFetches atomic.Int32
	shotBody    atomic.Value
}

func newFeedServer(t *testing.T) *feedServer {
	t.Helper()

	feed, err := os.ReadFile(filepath.Join("..", "..", "test", "fixtures", "templates", "templates.json"))
	if err != nil {
		t.Fatal(err)
	}

	server := &feedServer{}
	server.shotBody.Store(aPNG)
	server.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case server.failing.Load():
			http.Error(w, "down", http.StatusBadGateway)
		case r.URL.Path == "/templates.json":
			w.Write(feed)
		case strings.HasPrefix(r.URL.Path, "/templates/Summer/screenshot."):
			server.shotFetches.Add(1)
			w.Write(server.shotBody.Load().([]byte))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func (s *feedServer) feedURL() string { return s.URL + "/templates.json" }

type schedule struct {
	kind string
	when time.Time
}

type recordingScheduler struct{ calls []schedule }

func (r *recordingScheduler) EnqueueAt(_ context.Context, kind, _ string, _ any, when time.Time) (int64, error) {
	r.calls = append(r.calls, schedule{kind: kind, when: when})
	return 1, nil
}

func TestARefreshStoresTheFeedAndItReadsBack(t *testing.T) {
	h := newHarness(t)
	server := newFeedServer(t)
	catalogue := onboarding.NewCatalogue(h.db, server.feedURL(), h.stateDir, nil)

	if err := catalogue.Refresh(t.Context()); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	templates, err := catalogue.List(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) != 9 || templates[0].ID != "Bexer-Remix" {
		t.Fatalf("templates = %d, first %q", len(templates), templates[0].ID)
	}

	again, err := onboarding.NewCatalogue(h.db, server.feedURL(), h.stateDir, nil).List(t.Context())
	if err != nil || len(again) != 9 || again[8].Name != templates[8].Name {
		t.Errorf("a second catalogue read %d templates, %v", len(again), err)
	}
}

func TestAFailedRefreshKeepsTheLastGoodCopy(t *testing.T) {
	h := newHarness(t)
	server := newFeedServer(t)
	catalogue := onboarding.NewCatalogue(h.db, server.feedURL(), h.stateDir, nil)

	if err := catalogue.Refresh(t.Context()); err != nil {
		t.Fatal(err)
	}
	server.failing.Store(true)
	if err := catalogue.Refresh(t.Context()); err == nil {
		t.Fatal("a failing feed refreshed without error")
	}

	templates, err := catalogue.List(t.Context())
	if err != nil || len(templates) != 9 {
		t.Errorf("after a failed refresh: %d templates, %v", len(templates), err)
	}
}

func TestAFeedNeverReachedIsUnavailable(t *testing.T) {
	h := newHarness(t)
	server := newFeedServer(t)
	server.failing.Store(true)
	catalogue := onboarding.NewCatalogue(h.db, server.feedURL(), h.stateDir, nil)

	_ = catalogue.Refresh(t.Context())
	if _, err := catalogue.List(t.Context()); !errors.Is(err, onboarding.ErrGalleryUnavailable) {
		t.Errorf("List() error = %v, want ErrGalleryUnavailable", err)
	}
}

func TestACopyOfAnotherFeedIsNotServed(t *testing.T) {
	h := newHarness(t)
	server := newFeedServer(t)

	if err := onboarding.NewCatalogue(h.db, server.feedURL(), h.stateDir, nil).Refresh(t.Context()); err != nil {
		t.Fatal(err)
	}

	other := onboarding.NewCatalogue(h.db, server.URL+"/other.json", h.stateDir, nil)
	if _, err := other.List(t.Context()); !errors.Is(err, onboarding.ErrGalleryUnavailable) {
		t.Errorf("List() error = %v, want ErrGalleryUnavailable", err)
	}
}

func TestAGalleryThatIsOff(t *testing.T) {
	h := newHarness(t)
	catalogue := onboarding.NewCatalogue(h.db, "", h.stateDir, nil)
	scheduler := &recordingScheduler{}

	if _, err := catalogue.List(t.Context()); !errors.Is(err, onboarding.ErrGalleryOff) {
		t.Errorf("List() error = %v, want ErrGalleryOff", err)
	}
	if err := catalogue.Start(t.Context(), scheduler); err != nil || len(scheduler.calls) != 0 {
		t.Errorf("Start() scheduled %d refreshes, %v; want none", len(scheduler.calls), err)
	}
}

func TestTheRefreshJobSchedulesItsNextRun(t *testing.T) {
	h := newHarness(t)
	server := newFeedServer(t)
	catalogue := onboarding.NewCatalogue(h.db, server.feedURL(), h.stateDir, nil)

	cases := []struct {
		name    string
		failing bool
		want    time.Duration
	}{
		{"never fetched, failing", true, 5 * time.Minute},
		{"fetched", false, time.Hour},
		{"failing with a copy", true, time.Hour},
	}
	for _, c := range cases {
		server.failing.Store(c.failing)
		scheduler := &recordingScheduler{}
		var reported error
		before := time.Now()

		if err := catalogue.RefreshJob(scheduler, func(err error) { reported = err })(t.Context()); err != nil {
			t.Fatalf("%s: job error = %v", c.name, err)
		}
		if len(scheduler.calls) != 1 || scheduler.calls[0].kind != onboarding.TemplateRefreshJobKind {
			t.Fatalf("%s: scheduled %+v", c.name, scheduler.calls)
		}
		gap := scheduler.calls[0].when.Sub(before)
		if gap < c.want-time.Minute || gap > c.want+time.Minute {
			t.Errorf("%s: next run in %v, want about %v", c.name, gap, c.want)
		}
		if (reported != nil) != c.failing {
			t.Errorf("%s: reported %v", c.name, reported)
		}
	}
}

func TestStartingQueuesARefreshNow(t *testing.T) {
	h := newHarness(t)
	catalogue := onboarding.NewCatalogue(h.db, onboarding.DefaultTemplateFeed, h.stateDir, nil)
	scheduler := &recordingScheduler{}

	if err := catalogue.Start(t.Context(), scheduler); err != nil {
		t.Fatal(err)
	}
	if len(scheduler.calls) != 1 || time.Since(scheduler.calls[0].when) > time.Minute {
		t.Errorf("Start() scheduled %+v, want one refresh now", scheduler.calls)
	}
}

func TestAScreenshotIsFetchedOnceAndKept(t *testing.T) {
	h := newHarness(t)
	server := newFeedServer(t)
	catalogue := onboarding.NewCatalogue(h.db, server.feedURL(), h.stateDir, nil)
	if err := catalogue.Refresh(t.Context()); err != nil {
		t.Fatal(err)
	}

	for range 2 {
		data, kind, err := catalogue.Screenshot(t.Context(), "Summer")
		if err != nil {
			t.Fatalf("Screenshot() error = %v", err)
		}
		if kind != "image/png" || len(data) != len(aPNG) {
			t.Errorf("kind %q, %d bytes", kind, len(data))
		}
	}
	if got := server.shotFetches.Load(); got != 1 {
		t.Errorf("the screenshot was fetched %d times, want once", got)
	}
}

func TestAScreenshotThatIsNotAPicture(t *testing.T) {
	h := newHarness(t)
	server := newFeedServer(t)
	server.shotBody.Store([]byte("<html>not a picture</html>"))
	catalogue := onboarding.NewCatalogue(h.db, server.feedURL(), h.stateDir, nil)
	if err := catalogue.Refresh(t.Context()); err != nil {
		t.Fatal(err)
	}

	if _, _, err := catalogue.Screenshot(t.Context(), "Summer"); !errors.Is(err, onboarding.ErrNoScreenshot) {
		t.Errorf("Screenshot() error = %v, want ErrNoScreenshot", err)
	}
	if _, _, err := catalogue.Screenshot(t.Context(), "Nobody"); !errors.Is(err, onboarding.ErrNoSuchTemplate) {
		t.Errorf("an unknown template gave %v, want ErrNoSuchTemplate", err)
	}
}
