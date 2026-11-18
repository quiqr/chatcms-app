//go:build integration

package app_test

import (
	"io/fs"
	"testing"

	"github.com/quiqr/chatcms-app/internal/app"
	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/jobs"
	"github.com/quiqr/chatcms-app/internal/store"
)

func TestSaveReachesTheSharedVersionByJob(t *testing.T) {
	cfg := testConfig(t)
	logger := quietLogger()

	gitCfg := gitsync.Config{
		GitBin: gitBinary(t),
		System: gitsync.Identity{Name: "Quiqr ChatCMS", Email: "chatcms@example.org"},
	}
	sites := gitsync.NewSites(gitCfg, cfg.SitesDir(), nil)
	repository, err := sites.Clone(t.Context(), "korenaar", bareRemoteWithAPage(t, gitCfg))
	if err != nil {
		t.Fatal(err)
	}
	repository.Close()

	mine, err := sites.OpenFor(t.Context(), "korenaar", "anna")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mine.Save(t.Context(), gitsync.Save{
		Message: "Edit about",
		Author:  gitsync.Identity{Name: "Anna", Email: "anna@example.org"},
		Writes:  map[string][]byte{"content/about.md": []byte("anna\n")},
	}); err != nil {
		t.Fatal(err)
	}
	mine.Close()

	db, err := store.Open(t.Context(), cfg.DatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}

	queue := jobs.NewQueue(db)
	worker := jobs.NewWorker(queue, jobs.WorkerOptions{Logger: logger})
	app.RegisterShareJobForTest(worker, queue, sites, logger)
	app.RegisterSyncJobForTest(worker, sites, logger)

	if _, err := queue.Enqueue(t.Context(), app.ShareJobKind, "korenaar/anna", nil); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if did, err := worker.RunOnce(t.Context()); err != nil || !did {
			t.Fatalf("RunOnce() = %v, %v, want the share and then the sync to run", did, err)
		}
	}
	if done, _ := queue.CountByState(t.Context(), "done"); done != 2 {
		t.Fatalf("done = %d, want the share and the sync it queued", done)
	}

	shared, err := sites.Open(t.Context(), "korenaar")
	if err != nil {
		t.Fatal(err)
	}
	defer shared.Close()
	if data, err := fs.ReadFile(shared.FS(), "content/about.md"); err != nil || string(data) != "anna\n" {
		t.Errorf("shared version holds %q, %v, want anna's save", data, err)
	}
}
