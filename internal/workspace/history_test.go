package workspace_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/quiqr/chatcms-app/internal/runtime"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

type history struct {
	mu      sync.Mutex
	events  []workspace.Event
	begun   []workspace.Event
	samples []string
	forgot  []time.Time
}

func (h *history) Begin(_ context.Context, event workspace.Event) (int64, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.begun = append(h.begun, event)
	return int64(len(h.begun)), nil
}

func (h *history) Finish(_ context.Context, _ int64, event workspace.Event) error {
	return h.Record(context.Background(), event)
}

func (h *history) Record(_ context.Context, event workspace.Event) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = append(h.events, event)
	return nil
}

func (h *history) Sample(_ context.Context, at time.Time, starting, warm int) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.samples = append(h.samples, fmt.Sprintf("%s %d/%d", at.Format("15:04"), starting, warm))
	return nil
}

func (h *history) Forget(_ context.Context, before time.Time) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.forgot = append(h.forgot, before)
	return nil
}

func (h *history) kinds() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	var kinds []string
	for _, event := range h.events {
		kinds = append(kinds, string(event.Kind))
	}
	return kinds
}

func withHistory(t *testing.T, h *harness, most int, days int) *history {
	t.Helper()
	kept := &history{}
	return withHistoryAnd(t, h, kept, workspace.Settings{Most: most, Idle: 15 * time.Minute, History: kept, HistoryDays: days})
}

func withHistoryAnd(t *testing.T, h *harness, kept *history, settings workspace.Settings) *history {
	t.Helper()
	made, err := workspace.New(h.backend, h.store, h.copies, h.isAllowed, settings)
	if err != nil {
		t.Fatal(err)
	}
	made.PretendTimeIs(h.now)
	t.Cleanup(func() { made.Close(context.Background()) })
	h.w = made
	return kept
}

func TestStartsRefusalsAndFailuresAreRecorded(t *testing.T) {
	h := newHarness(t, 1)
	kept := withHistory(t, h, 1, 0)

	first := aKey()
	endpoint(t, h.w, first)
	if _, err := h.w.Open(t.Context(), aKey(), "u"); !errors.Is(err, workspace.ErrTooMany) {
		t.Fatalf("Open() = %v, want the cap", err)
	}
	h.w.Sleep(t.Context(), first)

	h.backend.Fail = workspace.ErrNotReady
	failing := aKey()
	h.w.Endpoint(t.Context(), failing, "u")

	if got := strings.Join(kept.kinds(), ","); got != "started,refused,failed" {
		t.Fatalf("recorded %s", got)
	}
	kept.mu.Lock()
	defer kept.mu.Unlock()
	if kept.events[0].Project != first.Project || kept.events[0].Backend == "" || kept.events[0].Millis < 0 {
		t.Fatalf("the start = %+v", kept.events[0])
	}
	if !strings.HasPrefix(kept.events[2].Reason, "did-not-start: ") || kept.events[2].Project != failing.Project {
		t.Fatalf("the failure = %+v", kept.events[2])
	}
}

func TestSamplesCountWhatRunsAndOldHistoryIsForgotten(t *testing.T) {
	h := newHarness(t, 8)
	kept := withHistory(t, h, 8, 30)
	endpoint(t, h.w, aKey())
	endpoint(t, h.w, aKey())

	h.w.SampleNow(t.Context())
	if len(kept.samples) != 1 || kept.samples[0] != "09:00 0/2" {
		t.Fatalf("samples = %v", kept.samples)
	}

	if err := h.w.CleanupNow(t.Context()); err != nil {
		t.Fatal(err)
	}
	h.w.ForgetHistoryNow(t.Context())
	if len(kept.forgot) != 1 || !kept.forgot[0].Equal(h.now().AddDate(0, 0, -30)) {
		t.Fatalf("forgot before %v", kept.forgot)
	}
}

func TestWithoutHistoryNothingBreaks(t *testing.T) {
	h := newHarness(t, 8)
	endpoint(t, h.w, aKey())
	h.w.SampleNow(t.Context())
	h.w.ForgetHistoryNow(t.Context())
}

func TestAStartIsKeptWithItsStepsLogAndHugo(t *testing.T) {
	h := newHarness(t, 8)
	kept := &history{}
	withHistoryAnd(t, h, kept, workspace.Settings{Most: 8, Idle: 15 * time.Minute, History: kept, HistoryDays: 90,
		RuntimeFor: func(context.Context, workspace.Key, string) (runtime.Resolved, error) {
			return runtime.Resolved{Preview: runtime.Command{Program: "/nix/store/h-hugo-0.101.0/bin/hugo"}, Hugo: &runtime.HugoChoice{Asked: "0.101.0", Program: "/nix/store/h-hugo-0.101.0/bin/hugo", Because: runtime.HugoPinned}}, nil
		}})
	h.backend.Says = []string{"Start building sites … ", "hugo v0.101.0-466fa43c16709b4483689930a4f9ac8add5c9f66+extended linux/amd64 BuildDate=unknown"}
	key := aKey()
	endpoint(t, h.w, key)

	kept.mu.Lock()
	defer kept.mu.Unlock()
	if len(kept.begun) != 1 || kept.begun[0].Person != key.Person || kept.begun[0].Kind != workspace.Begun {
		t.Fatalf("begun = %+v", kept.begun)
	}
	ended := kept.events[0]
	if ended.Kind != workspace.Started || len(ended.Steps) != 3 || ended.Person != key.Person {
		t.Fatalf("ended = %+v", ended)
	}
	log := strings.Join(ended.Log, "\n")
	for _, want := range []string{"backend fake", `hugo asked "0.101.0", chose /nix/store/h-hugo-0.101.0/bin/hugo (pinned)`} {
		if !strings.Contains(log, want) {
			t.Errorf("the log lacks %q:\n%s", want, log)
		}
	}
	want := workspace.HugoFacts{Asked: "0.101.0", Program: "/nix/store/h-hugo-0.101.0/bin/hugo", Because: "pinned", Printed: "hugo v0.101.0-466fa43c16709b4483689930a4f9ac8add5c9f66+extended linux/amd64 BuildDate=unknown"}
	if ended.Hugo == nil || *ended.Hugo != want {
		t.Errorf("hugo = %+v, want %+v", ended.Hugo, want)
	}
}
