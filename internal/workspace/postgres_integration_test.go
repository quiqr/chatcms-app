//go:build integration

package workspace_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/store"
	"github.com/quiqr/chatcms-app/internal/testsupport"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

func database(t *testing.T) *store.DB {
	t.Helper()

	db, err := store.Open(t.Context(), testsupport.Postgres(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	if err := db.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}
	return db
}

func keysIn(t *testing.T, db *store.DB, count int) []workspace.Key {
	t.Helper()

	organisation := uuid.New()
	project := uuid.New()
	if _, err := db.Pool().Exec(t.Context(), `INSERT INTO organisations (id, slug, name) VALUES ($1, $2, 'Studio')`, organisation, "studio-"+organisation.String()[:8]); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Pool().Exec(t.Context(), `INSERT INTO projects (id, organisation_id, slug, name) VALUES ($1, $2, 'site', 'Site')`, project, organisation); err != nil {
		t.Fatal(err)
	}

	keys := make([]workspace.Key, 0, count)
	for i := range count {
		person := uuid.New()
		if _, err := db.Pool().Exec(t.Context(), `INSERT INTO people (id, email) VALUES ($1, $2)`, person, fmt.Sprintf("p%d-%s@example.org", i, person)); err != nil {
			t.Fatal(err)
		}
		keys = append(keys, workspace.Key{Person: person, Project: project})
	}
	return keys
}

func TestTheLastPlaceIsTakenOnce(t *testing.T) {
	db := database(t)
	places := workspace.NewPostgres(db)
	keys := keysIn(t, db, 6)
	now := time.Now()

	if _, err := places.Reserve(t.Context(), keys[0], "local", 2, now); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	results := make(chan error, 5)
	for _, key := range keys[1:] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := places.Reserve(t.Context(), key, "local", 2, now)
			results <- err
		}()
	}
	wg.Wait()
	close(results)

	taken, refused := 0, 0
	for err := range results {
		switch {
		case err == nil:
			taken++
		case errors.Is(err, workspace.ErrTooMany):
			refused++
		default:
			t.Errorf("Reserve() error = %v", err)
		}
	}
	if taken != 1 || refused != 4 {
		t.Errorf("taken = %d, refused = %d, want the last place taken once", taken, refused)
	}
}

func TestPostgresKeepsTheLifecycle(t *testing.T) {
	db := database(t)
	places := workspace.NewPostgres(db)
	key := keysIn(t, db, 1)[0]
	at := time.Date(2026, 11, 18, 9, 0, 0, 0, time.UTC)

	record, err := places.Reserve(t.Context(), key, "local", 8, at)
	if err != nil || record.State != workspace.Starting || record.Key != key {
		t.Fatalf("Reserve() = %+v, %v", record, err)
	}
	again, err := places.Reserve(t.Context(), key, "local", 1, at)
	if err != nil || again.ID != record.ID {
		t.Fatalf("Reserve() again = %+v, %v, want the same row, its own place not counted", again, err)
	}

	if err := places.Warm(t.Context(), record.ID, "pid:1:hugo"); err != nil {
		t.Fatal(err)
	}
	busy, err := places.Busy(t.Context())
	if err != nil || len(busy) != 1 || busy[0].Handle != "pid:1:hugo" || busy[0].State != workspace.Warm {
		t.Fatalf("Busy() = %+v, %v", busy, err)
	}

	later := at.Add(time.Hour)
	if err := places.Touch(t.Context(), record.ID, later); err != nil {
		t.Fatal(err)
	}
	if err := places.Touch(t.Context(), record.ID, at); err != nil {
		t.Fatal(err)
	}
	got, err := places.Get(t.Context(), key)
	if err != nil || !got.LastUsed.Equal(later) {
		t.Fatalf("Get() = %+v, %v, want the later use kept", got, err)
	}

	if err := places.Sleep(t.Context(), record.ID, "did-not-start"); err != nil {
		t.Fatal(err)
	}
	sleeping, err := places.Sleeping(t.Context())
	if err != nil || len(sleeping) != 1 || sleeping[0].Reason != "did-not-start" || sleeping[0].Handle != "" {
		t.Fatalf("Sleeping() = %+v, %v", sleeping, err)
	}

	if err := places.Remove(t.Context(), record.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := places.Get(t.Context(), key); !errors.Is(err, workspace.ErrNotFound) {
		t.Errorf("Get() after Remove() error = %v, want ErrNotFound", err)
	}
}

func TestAWorkspaceGoesWithItsPerson(t *testing.T) {
	db := database(t)
	places := workspace.NewPostgres(db)
	key := keysIn(t, db, 1)[0]
	if _, err := places.Reserve(t.Context(), key, "local", 8, time.Now()); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Pool().Exec(t.Context(), `DELETE FROM people WHERE id = $1`, key.Person); err != nil {
		t.Fatal(err)
	}
	if _, err := places.Get(t.Context(), key); !errors.Is(err, workspace.ErrNotFound) {
		t.Errorf("Get() error = %v, want the workspace gone with its person", err)
	}
}
