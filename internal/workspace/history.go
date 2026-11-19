package workspace

import (
	"context"
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
	Started EventKind = "started"
	Refused EventKind = "refused"
	Failed  EventKind = "failed"
)

type Event struct {
	At      time.Time
	Kind    EventKind
	Backend string
	Project uuid.UUID
	Millis  int
	Reason  string
}

type History interface {
	Record(ctx context.Context, event Event) error
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
