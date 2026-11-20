package workspace

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultHistoryDays = 90
	sampleEvery        = 5 * time.Minute
)

type EventKind string

const (
	Begun   EventKind = "starting"
	Started EventKind = "started"
	Refused EventKind = "refused"
	Failed  EventKind = "failed"
)

type Event struct {
	At      time.Time
	Kind    EventKind
	Backend string
	Project uuid.UUID
	Person  uuid.UUID
	Millis  int
	Reason  string
	Steps   []StepReport
	Log     []string
	Hugo    *HugoFacts
}

type HugoFacts struct {
	Asked   string `json:"asked"`
	Program string `json:"program"`
	Because string `json:"because"`
	Printed string `json:"printed"`
}

type History interface {
	Record(ctx context.Context, event Event) error
	Begin(ctx context.Context, event Event) (int64, error)
	Finish(ctx context.Context, id int64, event Event) error
	Sample(ctx context.Context, at time.Time, starting, warm int) error
	Forget(ctx context.Context, before time.Time) error
}

func firstLine(text string, most int) string {
	line, _, _ := strings.Cut(text, "\n")
	if len(line) > most {
		line = line[:most]
	}
	return line
}

var hugoSays = regexp.MustCompile(`hugo v\d+\.\d+\S*`)

func printedBy(lines []string) string {
	for _, line := range lines {
		if found := hugoSays.FindString(line); found != "" {
			return firstLine(strings.TrimSpace(line), 300)
		}
	}
	return ""
}

func (w *Workspaces) beginStart(event Event) int64 {
	if w.settings.History == nil {
		return 0
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	id, err := w.settings.History.Begin(ctx, event)
	if err != nil {
		w.settings.Logger.Warn("recording a workspace start failed", "error", err)
	}
	return id
}

func (w *Workspaces) finishStart(id int64, event Event) {
	if w.settings.History == nil || id == 0 {
		w.record(event)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.settings.History.Finish(ctx, id, event); err != nil {
		w.settings.Logger.Warn("recording how a workspace start ended failed", "error", err)
	}
}

func (w *Workspaces) record(event Event) {
	if w.settings.History == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.settings.History.Record(ctx, event); err != nil {
		w.settings.Logger.Warn("recording workspace history failed", "error", err)
	}
}

func (w *Workspaces) SampleNow(ctx context.Context) {
	if w.settings.History == nil {
		return
	}
	w.mu.Lock()
	starting, warm := 0, 0
	for _, held := range w.live {
		switch held.state {
		case Starting:
			starting++
		case Warm:
			warm++
		}
	}
	at := w.now().Truncate(sampleEvery)
	w.mu.Unlock()
	if err := w.settings.History.Sample(ctx, at, starting, warm); err != nil {
		w.settings.Logger.Warn("sampling workspaces failed", "error", err)
	}
}

func (w *Workspaces) ForgetHistoryNow(ctx context.Context) {
	if w.settings.History == nil {
		return
	}
	before := w.clock().AddDate(0, 0, -w.settings.HistoryDays)
	if err := w.settings.History.Forget(ctx, before); err != nil {
		w.settings.Logger.Warn("forgetting old workspace history failed", "error", err)
	}
}
