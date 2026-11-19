package workspace

import (
	"context"
	"errors"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms-app/internal/store"
)

type Counts struct {
	Starting int `json:"starting"`
	Warm     int `json:"warm"`
	Sleeping int `json:"sleeping"`
}

func (c *Counts) add(state string, n int) {
	switch State(state) {
	case Starting:
		c.Starting += n
	case Warm:
		c.Warm += n
	case Sleeping:
		c.Sleeping += n
	}
}

type OrganisationUse struct {
	Slug   string `json:"slug"`
	Name   string `json:"name"`
	Counts Counts `json:"counts"`
}

type ProjectUse struct {
	ID           uuid.UUID `json:"-"`
	Organisation string    `json:"organisation"`
	Slug         string    `json:"slug"`
	Name         string    `json:"name"`
	Counts       Counts    `json:"counts"`
	DiskBytes    int64     `json:"diskBytes"`
	DiskKnown    bool      `json:"diskKnown"`
}

type ReasonCount struct {
	Reason string `json:"reason"`
	Times  int    `json:"times"`
}

type DayPeak struct {
	Day      string `json:"day"`
	Warm     int    `json:"warm"`
	Starting int    `json:"starting"`
}

type Statistics struct {
	Backends      []string          `json:"backends"`
	Limit         int               `json:"limit"`
	Now           Counts            `json:"now"`
	Organisations []OrganisationUse `json:"organisations"`
	Projects      []ProjectUse      `json:"projects"`
	Days          int               `json:"days"`
	Starts        int               `json:"starts"`
	MedianMillis  int               `json:"medianMillis"`
	SlowMillis    int               `json:"slowMillis"`
	Refused       int               `json:"refused"`
	Failed        []ReasonCount     `json:"failed"`
	Peaks         []DayPeak         `json:"peaks"`
	HistoryDays   int               `json:"historyDays"`
}

func (p *Postgres) Record(ctx context.Context, event Event) error {
	var millis *int32
	if event.Kind == Started {
		value := int32(event.Millis)
		millis = &value
	}
	project := pgtype.UUID{}
	if event.Project != uuid.Nil {
		project = pg(event.Project)
	}
	return p.db.Queries.RecordWorkspaceEvent(ctx, store.RecordWorkspaceEventParams{
		At:        pgTime(event.At),
		Kind:      string(event.Kind),
		Backend:   event.Backend,
		ProjectID: project,
		Millis:    millis,
		Reason:    event.Reason,
	})
}

func (p *Postgres) Sample(ctx context.Context, at time.Time, starting, warm int) error {
	return p.db.Queries.SampleWorkspaces(ctx, store.SampleWorkspacesParams{At: pgTime(at), Starting: int32(starting), Warm: int32(warm)})
}

func (p *Postgres) Forget(ctx context.Context, before time.Time) error {
	if err := p.db.Queries.ForgetWorkspaceEvents(ctx, pgTime(before)); err != nil {
		return err
	}
	return p.db.Queries.ForgetWorkspaceSamples(ctx, pgTime(before))
}

func (p *Postgres) Statistics(ctx context.Context, since time.Time) (Statistics, error) {
	stats := Statistics{Organisations: []OrganisationUse{}, Projects: []ProjectUse{}, Failed: []ReasonCount{}, Peaks: []DayPeak{}, Backends: []string{}}
	rows, err := p.db.Queries.CountWorkspacesByState(ctx)
	if err != nil {
		return Statistics{}, err
	}
	organisations := map[string]int{}
	projects := map[uuid.UUID]int{}
	backends := map[string]bool{}
	for _, row := range rows {
		n := int(row.Workspaces)
		stats.Now.add(row.State, n)
		if row.State != string(Sleeping) && !backends[row.Backend] {
			backends[row.Backend] = true
			stats.Backends = append(stats.Backends, row.Backend)
		}
		at, ok := organisations[row.OrganisationSlug]
		if !ok {
			at = len(stats.Organisations)
			organisations[row.OrganisationSlug] = at
			stats.Organisations = append(stats.Organisations, OrganisationUse{Slug: row.OrganisationSlug, Name: row.OrganisationName})
		}
		stats.Organisations[at].Counts.add(row.State, n)
		id := uuid.UUID(row.ProjectID.Bytes)
		at, ok = projects[id]
		if !ok {
			at = len(stats.Projects)
			projects[id] = at
			stats.Projects = append(stats.Projects, ProjectUse{ID: id, Organisation: row.OrganisationSlug, Slug: row.ProjectSlug, Name: row.ProjectName})
		}
		stats.Projects[at].Counts.add(row.State, n)
	}

	starts, err := p.db.Queries.WorkspaceStartsSince(ctx, pgTime(since))
	if err != nil {
		return Statistics{}, err
	}
	stats.Starts = int(starts.Starts)
	stats.MedianMillis = int(math.Round(starts.MedianMillis))
	stats.SlowMillis = int(math.Round(starts.SlowMillis))

	events, err := p.db.Queries.CountWorkspaceEventsSince(ctx, pgTime(since))
	if err != nil {
		return Statistics{}, err
	}
	for _, event := range events {
		if event.Kind == string(Refused) {
			stats.Refused += int(event.Times)
			continue
		}
		stats.Failed = append(stats.Failed, ReasonCount{Reason: event.Reason, Times: int(event.Times)})
	}

	peaks, err := p.db.Queries.WorkspacePeaksSince(ctx, pgTime(since))
	if err != nil {
		return Statistics{}, err
	}
	for _, peak := range peaks {
		stats.Peaks = append(stats.Peaks, DayPeak{Day: peak.Day.Time.Format("2006-01-02"), Warm: int(peak.Warm), Starting: int(peak.Starting)})
	}
	return stats, nil
}

var errTooLong = errors.New("workspace: measuring took too long")

func DiskOf(dir string, within time.Duration) (int64, bool) {
	deadline := time.Now().Add(within)
	var total int64
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if time.Now().After(deadline) {
			return errTooLong
		}
		if entry.Type().IsRegular() {
			info, err := entry.Info()
			if err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	if err != nil {
		return 0, false
	}
	if _, statErr := os.Stat(dir); statErr != nil {
		return 0, true
	}
	return total, true
}
