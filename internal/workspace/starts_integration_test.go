//go:build integration

package workspace_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms-app/internal/store"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

func TestAStartIsKeptAndTheProjectLogMergesItWithBuildsAndPublications(t *testing.T) {
	db := database(t)
	history := workspace.NewPostgres(db)
	keys := keysIn(t, db, 1)
	project := pgtype.UUID{Bytes: keys[0].Project, Valid: true}
	now := time.Now().UTC().Truncate(time.Millisecond)

	id, err := history.Begin(t.Context(), workspace.Event{At: now.Add(-time.Minute), Kind: workspace.Begun, Backend: "gvisor", Project: keys[0].Project, Person: keys[0].Person})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Queries.StartBuild(t.Context(), store.StartBuildParams{ProjectID: project, Reason: "publish"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Queries.StartPublish(t.Context(), store.StartPublishParams{ProjectID: project, PersonID: pgtype.UUID{Bytes: keys[0].Person, Valid: true}, ChangeCount: 2, Summary: []byte("[]")}); err != nil {
		t.Fatal(err)
	}

	running, err := db.Queries.GetWorkspaceStart(t.Context(), store.GetWorkspaceStartParams{ID: id, ProjectID: project})
	if err != nil || running.Kind != "starting" || running.Who == "" {
		t.Fatalf("running start = %+v, %v", running, err)
	}

	if err := history.Finish(t.Context(), id, workspace.Event{
		At: now.Add(-time.Minute), Kind: workspace.Failed, Millis: 4100, Reason: "did-not-start: no answer",
		Steps: []workspace.StepReport{{Step: workspace.StepBuilding, State: workspace.StepFailed}},
		Log:   []string{"backend gvisor", "Error: template failed"},
		Hugo:  &workspace.HugoFacts{Asked: "0.101.0", Program: "/nix/store/h/bin/hugo", Because: "pinned", Printed: "hugo v0.101.0"},
	}); err != nil {
		t.Fatal(err)
	}
	kept, err := db.Queries.GetWorkspaceStart(t.Context(), store.GetWorkspaceStartParams{ID: id, ProjectID: project})
	if err != nil {
		t.Fatal(err)
	}
	var hugo workspace.HugoFacts
	if json.Unmarshal(kept.Hugo, &hugo) != nil || hugo.Asked != "0.101.0" || hugo.Printed != "hugo v0.101.0" {
		t.Errorf("hugo = %s", kept.Hugo)
	}
	if kept.Kind != "failed" || kept.Log != "backend gvisor\nError: template failed" || kept.Millis == nil || *kept.Millis != 4100 || !kept.FinishedAt.Valid {
		t.Errorf("kept = %+v", kept)
	}

	rows, err := db.Queries.ProjectLog(t.Context(), store.ProjectLogParams{Project: project, Most: 10})
	if err != nil {
		t.Fatal(err)
	}
	var kinds []string
	for _, row := range rows {
		kinds = append(kinds, row.Kind)
	}
	if len(kinds) != 3 || kinds[2] != "start" {
		t.Fatalf("log = %v, want the start last as the oldest", kinds)
	}
	only := "start"
	rows, _ = db.Queries.ProjectLog(t.Context(), store.ProjectLogParams{Project: project, OfKind: &only, Most: 10})
	if len(rows) != 1 || rows[0].State != "failed" || rows[0].What != "gvisor" || rows[0].Who == "" {
		t.Errorf("starts only = %+v", rows)
	}
	rows, _ = db.Queries.ProjectLog(t.Context(), store.ProjectLogParams{Project: project, Before: pgtype.Timestamptz{Time: now.Add(-30 * time.Second), Valid: true}, Most: 10})
	if len(rows) != 1 || rows[0].Kind != "start" {
		t.Errorf("before = %+v", rows)
	}
}

func TestAStartLeftRunningIsMarkedInterrupted(t *testing.T) {
	db := database(t)
	history := workspace.NewPostgres(db)
	keys := keysIn(t, db, 1)
	id, err := history.Begin(t.Context(), workspace.Event{At: time.Now(), Kind: workspace.Begun, Backend: "local", Project: keys[0].Project, Person: keys[0].Person})
	if err != nil {
		t.Fatal(err)
	}
	if err := history.Interrupt(t.Context()); err != nil {
		t.Fatal(err)
	}
	kept, err := db.Queries.GetWorkspaceStart(t.Context(), store.GetWorkspaceStartParams{ID: id, ProjectID: pgtype.UUID{Bytes: keys[0].Project, Valid: true}})
	if err != nil || kept.Kind != "interrupted" || !kept.FinishedAt.Valid {
		t.Errorf("kept = %+v, %v", kept, err)
	}
}
