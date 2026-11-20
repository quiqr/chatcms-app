package app

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/policy"
	"github.com/quiqr/chatcms-app/internal/store"
	"github.com/quiqr/chatcms-app/internal/tenancy"
	"github.com/quiqr/chatcms-app/internal/workflows"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

const logPage = 50

var logKinds = map[string]bool{"start": true, "build": true, "publish": true}

type projectLog struct {
	queries *store.Queries
	tenancy *tenancy.Store
	sources publishSources
}

func (l projectLog) reach(ctx context.Context, who httpapi.Actor, organisation, project string) (uuid.UUID, error) {
	if who.Anonymous || who.ID == "" {
		return uuid.Nil, httpapi.ErrNotAllowed
	}
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return uuid.Nil, httpapi.ErrNotAllowed
	}
	actor, err := l.tenancy.ActorFor(ctx, personID)
	if err != nil {
		return uuid.Nil, err
	}
	found, err := l.tenancy.ProjectBySlug(ctx, organisation, project)
	if errors.Is(err, tenancy.ErrNotFound) {
		return uuid.Nil, httpapi.ErrMissing
	}
	if err != nil {
		return uuid.Nil, err
	}
	resource, err := l.sources.Resource(ctx, found.ID)
	if err != nil {
		return uuid.Nil, err
	}
	if !policy.Can(actor, policy.ActionReadContent, resource) {
		return uuid.Nil, httpapi.ErrMissing
	}
	if !policy.Can(actor, policy.ActionSeeTechnicalDetail, resource) {
		return uuid.Nil, httpapi.ErrNotAllowed
	}
	return found.ID, nil
}

func stamp(at pgtype.Timestamptz) string {
	if !at.Valid {
		return ""
	}
	return at.Time.UTC().Format(time.RFC3339Nano)
}

func millisOf(value *int32) *int {
	if value == nil {
		return nil
	}
	millis := int(*value)
	return &millis
}

func (l projectLog) ProjectLog(ctx context.Context, who httpapi.Actor, organisation, project, kind, before string) (httpapi.LogView, error) {
	projectID, err := l.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.LogView{}, err
	}
	params := store.ProjectLogParams{Project: pgUUID(projectID), Most: logPage + 1}
	if logKinds[kind] {
		params.OfKind = &kind
	}
	if at, err := time.Parse(time.RFC3339Nano, before); err == nil {
		params.Before = pgtype.Timestamptz{Time: at, Valid: true}
	}
	rows, err := l.queries.ProjectLog(ctx, params)
	if err != nil {
		return httpapi.LogView{}, err
	}
	view := httpapi.LogView{Entries: []httpapi.LogEntryView{}}
	for index, row := range rows {
		if index == logPage {
			view.More = true
			break
		}
		view.Entries = append(view.Entries, httpapi.LogEntryView{
			Entry: row.Kind + ":" + row.ID, Kind: row.Kind, At: stamp(row.At), Finished: stamp(row.FinishedAt),
			State: row.State, What: row.What, Who: row.Who, Millis: millisOf(row.Millis),
		})
	}
	return view, nil
}

func (l projectLog) LogEntry(ctx context.Context, who httpapi.Actor, organisation, project, entry string) (httpapi.LogDetailView, error) {
	projectID, err := l.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.LogDetailView{}, err
	}
	kind, id, _ := strings.Cut(entry, ":")
	var view httpapi.LogDetailView
	switch kind {
	case "start":
		view, err = l.start(ctx, projectID, id)
	case "build":
		view, err = l.build(ctx, projectID, id)
	case "publish":
		view, err = l.publish(ctx, projectID, id)
	default:
		return httpapi.LogDetailView{}, httpapi.ErrMissing
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return httpapi.LogDetailView{}, httpapi.ErrMissing
	}
	if err == nil {
		view.Entry, view.Kind = entry, kind
	}
	return view, err
}

func (l projectLog) start(ctx context.Context, projectID uuid.UUID, id string) (httpapi.LogDetailView, error) {
	number, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return httpapi.LogDetailView{}, pgx.ErrNoRows
	}
	row, err := l.queries.GetWorkspaceStart(ctx, store.GetWorkspaceStartParams{ID: number, ProjectID: pgUUID(projectID)})
	if err != nil {
		return httpapi.LogDetailView{}, err
	}
	view := httpapi.LogDetailView{
		LogEntryView: httpapi.LogEntryView{At: stamp(row.At), Finished: stamp(row.FinishedAt), State: row.Kind, What: row.Backend, Who: row.Who, Millis: millisOf(row.Millis)},
		Steps:        []httpapi.LogStepView{},
		Log:          row.Log,
		Reason:       row.Reason,
	}
	var steps []workspace.StepReport
	if json.Unmarshal(row.Steps, &steps) == nil {
		for _, step := range steps {
			view.Steps = append(view.Steps, httpapi.LogStepView{Step: string(step.Step), State: string(step.State)})
		}
	}
	if len(row.Hugo) > 0 {
		var facts httpapi.HugoFactsView
		if json.Unmarshal(row.Hugo, &facts) == nil {
			view.Hugo = &facts
		}
	}
	return view, nil
}

func parseUUID(id string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return pgtype.UUID{}, pgx.ErrNoRows
	}
	return pgUUID(parsed), nil
}

func (l projectLog) build(ctx context.Context, projectID uuid.UUID, id string) (httpapi.LogDetailView, error) {
	key, err := parseUUID(id)
	if err != nil {
		return httpapi.LogDetailView{}, err
	}
	row, err := l.queries.GetBuild(ctx, key)
	if err != nil {
		return httpapi.LogDetailView{}, err
	}
	if uuid.UUID(row.ProjectID.Bytes) != projectID {
		return httpapi.LogDetailView{}, pgx.ErrNoRows
	}
	view := httpapi.LogDetailView{
		LogEntryView: httpapi.LogEntryView{At: stamp(row.StartedAt), Finished: stamp(row.FinishedAt), State: row.State, What: row.Reason},
		Steps:        []httpapi.LogStepView{},
		Log:          row.Log,
	}
	if row.ExitCode != nil {
		exit := int(*row.ExitCode)
		view.Exit = &exit
	}
	return view, nil
}

func (l projectLog) publish(ctx context.Context, projectID uuid.UUID, id string) (httpapi.LogDetailView, error) {
	key, err := parseUUID(id)
	if err != nil {
		return httpapi.LogDetailView{}, err
	}
	row, err := l.queries.GetPublish(ctx, key)
	if err != nil {
		return httpapi.LogDetailView{}, err
	}
	if uuid.UUID(row.ProjectID.Bytes) != projectID {
		return httpapi.LogDetailView{}, pgx.ErrNoRows
	}
	view := httpapi.LogDetailView{
		LogEntryView: httpapi.LogEntryView{At: stamp(row.StartedAt), Finished: stamp(row.FinishedAt), State: row.State, What: strconv.Itoa(int(row.ChangeCount))},
		Steps:        []httpapi.LogStepView{},
		Log:          row.Detail,
		Reason:       row.Trouble,
	}
	var steps []workflows.StepOutcome
	if json.Unmarshal(row.Steps, &steps) == nil {
		for _, step := range steps {
			view.Steps = append(view.Steps, httpapi.LogStepView{Step: step.Title, State: string(step.State), Detail: step.Detail})
		}
	}
	return view, nil
}

var _ httpapi.ProjectLogService = projectLog{}
