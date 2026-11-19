//go:build integration

package workspace_test

import (
	"testing"
	"time"

	"github.com/quiqr/chatcms-app/internal/workspace"
)

func TestStatisticsCountNowAndTheDaysBefore(t *testing.T) {
	db := database(t)
	store := workspace.NewPostgres(db)
	keys := keysIn(t, db, 3)
	now := time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)

	warm, err := store.Reserve(t.Context(), keys[0], "gvisor", 8, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Warm(t.Context(), warm.ID, "ws1"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Reserve(t.Context(), keys[1], "gvisor", 8, now); err != nil {
		t.Fatal(err)
	}
	asleep, err := store.Reserve(t.Context(), keys[2], "gvisor", 8, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Sleep(t.Context(), asleep.ID, ""); err != nil {
		t.Fatal(err)
	}

	project := keys[0].Project
	for _, event := range []workspace.Event{
		{At: now.Add(-time.Hour), Kind: workspace.Started, Backend: "gvisor", Project: project, Millis: 1000},
		{At: now.Add(-2 * time.Hour), Kind: workspace.Started, Backend: "gvisor", Project: project, Millis: 3000},
		{At: now.Add(-3 * time.Hour), Kind: workspace.Started, Backend: "gvisor", Project: project, Millis: 2000},
		{At: now.Add(-time.Hour), Kind: workspace.Refused, Backend: "gvisor", Project: project},
		{At: now.Add(-time.Hour), Kind: workspace.Failed, Backend: "gvisor", Project: project, Reason: "did-not-start: no answer"},
		{At: now.Add(-2 * time.Hour), Kind: workspace.Failed, Backend: "gvisor", Project: project, Reason: "did-not-start: no answer"},
		{At: now.AddDate(0, 0, -10), Kind: workspace.Started, Backend: "gvisor", Project: project, Millis: 99000},
	} {
		if err := store.Record(t.Context(), event); err != nil {
			t.Fatal(err)
		}
	}
	for at, warm := range map[time.Time]int{now.Add(-time.Hour): 3, now.Add(-30 * time.Minute): 5, now.AddDate(0, 0, -1): 2} {
		if err := store.Sample(t.Context(), at, 1, warm); err != nil {
			t.Fatal(err)
		}
	}

	stats, err := store.Statistics(t.Context(), now.AddDate(0, 0, -7))
	if err != nil {
		t.Fatal(err)
	}
	if stats.Now != (workspace.Counts{Starting: 1, Warm: 1, Sleeping: 1}) {
		t.Fatalf("now = %+v", stats.Now)
	}
	if len(stats.Organisations) != 1 || stats.Organisations[0].Counts != stats.Now {
		t.Fatalf("organisations = %+v", stats.Organisations)
	}
	if len(stats.Projects) != 1 || stats.Projects[0].ID != project || stats.Projects[0].Name != "Site" {
		t.Fatalf("projects = %+v", stats.Projects)
	}
	if len(stats.Backends) != 1 || stats.Backends[0] != "gvisor" {
		t.Fatalf("backends = %v", stats.Backends)
	}
	if stats.Starts != 3 || stats.MedianMillis != 2000 || stats.SlowMillis != 2800 {
		t.Fatalf("starts = %d, median %d, slow %d", stats.Starts, stats.MedianMillis, stats.SlowMillis)
	}
	if stats.Refused != 1 || len(stats.Failed) != 1 || stats.Failed[0] != (workspace.ReasonCount{Reason: "did-not-start: no answer", Times: 2}) {
		t.Fatalf("refused %d, failed %+v", stats.Refused, stats.Failed)
	}
	if len(stats.Peaks) != 2 || stats.Peaks[1] != (workspace.DayPeak{Day: "2026-11-19", Warm: 5, Starting: 1}) {
		t.Fatalf("peaks = %+v", stats.Peaks)
	}

	if err := store.Forget(t.Context(), now.AddDate(0, 0, -7)); err != nil {
		t.Fatal(err)
	}
	all, err := store.Statistics(t.Context(), now.AddDate(0, 0, -90))
	if err != nil {
		t.Fatal(err)
	}
	if all.Starts != 3 {
		t.Fatalf("after forgetting, starts = %d, want the old one gone", all.Starts)
	}
}
