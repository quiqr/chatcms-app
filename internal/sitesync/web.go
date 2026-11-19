package sitesync

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/policy"
	"github.com/quiqr/chatcms-app/internal/tenancy"
)

// Web answers the HTTP layer's questions.
type Web struct {
	service *Service
	tenancy *tenancy.Store
}

func NewWeb(service *Service, tenancyStore *tenancy.Store) (*Web, error) {
	if service == nil || tenancyStore == nil {
		return nil, ErrIncomplete
	}
	return &Web{service: service, tenancy: tenancyStore}, nil
}

func (w *Web) reach(ctx context.Context, who httpapi.Actor, organisation, project string) (policy.Actor, uuid.UUID, error) {
	if who.Anonymous || who.ID == "" {
		return policy.Actor{}, uuid.Nil, httpapi.ErrNotAllowed
	}
	personID, err := uuid.Parse(who.ID)
	if err != nil {
		return policy.Actor{}, uuid.Nil, httpapi.ErrNotAllowed
	}

	actor, err := w.tenancy.ActorFor(ctx, personID)
	if err != nil {
		return policy.Actor{}, uuid.Nil, err
	}

	found, err := w.tenancy.ProjectBySlug(ctx, organisation, project)
	if errors.Is(err, tenancy.ErrNotFound) {
		return policy.Actor{}, uuid.Nil, httpapi.ErrMissing
	}
	if err != nil {
		return policy.Actor{}, uuid.Nil, err
	}
	return actor, found.ID, nil
}

func (w *Web) SiteStanding(ctx context.Context, who httpapi.Actor, organisation, project string) (httpapi.StandingView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.StandingView{}, err
	}

	standing, err := w.service.Standing(ctx, actor, id)
	if err != nil {
		return httpapi.StandingView{}, err
	}
	return standingView(standing), nil
}

func (w *Web) BringChanges(ctx context.Context, who httpapi.Actor, organisation, project string) (httpapi.StandingView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.StandingView{}, err
	}

	standing, err := w.service.Bring(ctx, actor, id)
	if err != nil {
		return httpapi.StandingView{}, err
	}
	return standingView(standing), nil
}

func (w *Web) AnswerChanges(ctx context.Context, who httpapi.Actor, organisation, project string, asked []httpapi.AnswerRequest) (httpapi.StandingView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.StandingView{}, err
	}

	answers := make([]Answer, 0, len(asked))
	for _, one := range asked {
		answer := Answer{File: one.File, Whole: Choice(one.Whole), Fields: map[string]Choice{}}
		for field, choice := range one.Fields {
			answer.Fields[field] = Choice(choice)
		}
		if len(one.Combined) > 0 {
			answer.Combined = map[string]string{}
			for field, text := range one.Combined {
				answer.Combined[field] = text
			}
		}
		answers = append(answers, answer)
	}

	standing, err := w.service.Answered(ctx, actor, id, answers)
	if err != nil {
		return httpapi.StandingView{}, err
	}
	return standingView(standing), nil
}

func (w *Web) AbandonChanges(ctx context.Context, who httpapi.Actor, organisation, project string) (httpapi.StandingView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.StandingView{}, err
	}

	standing, err := w.service.Abandon(ctx, actor, id)
	if err != nil {
		return httpapi.StandingView{}, err
	}
	return standingView(standing), nil
}

// said is how each state reads to a content creator. The words are the ones
// docs/design/DESIGN.md section 5.10 uses, and none of them is a git word.
var inWords = map[State]string{
	Saved:   "All changes saved",
	Saving:  "Saving…",
	Offline: "Offline · saved on this device",
	Arrived: "Someone else updated this site",
	Asking:  "Two people changed the same thing",
}

func standingView(standing Standing) httpapi.StandingView {
	view := httpapi.StandingView{
		State:       string(standing.State),
		Said:        inWords[standing.State],
		Unpublished: standing.Unpublished,
		ArrivedFrom: standing.ArrivedFrom,
	}

	if standing.Technical != nil {
		view.Technical = &httpapi.TechnicalView{
			Branch: standing.Technical.Line,
			Remote: standing.Technical.Remote,
			Head:   standing.Technical.Head,
		}
	}

	for _, page := range standing.Conflicts {
		shown := httpapi.PageConflictView{
			Page:      page.Page,
			Title:     page.Title,
			File:      page.File,
			Who:       page.Who,
			WholeFile: page.WholeFile,
		}
		for _, field := range page.Fields {
			entry := httpapi.FieldConflictView{
				Path:   field.Path,
				Name:   field.Name,
				Mine:   field.Mine,
				Theirs: field.Theirs,
			}
			for _, word := range field.Words {
				entry.Words = append(entry.Words, httpapi.WordView{Part: string(word.Part), Text: word.Text})
			}
			shown.Fields = append(shown.Fields, entry)
		}
		view.Conflicts = append(view.Conflicts, shown)
	}
	return view
}

var _ httpapi.SyncService = (*Web)(nil)

func (w *Web) CombineChanges(ctx context.Context, who httpapi.Actor, organisation, project, file, field string) (httpapi.CombinedView, error) {
	actor, id, err := w.reach(ctx, who, organisation, project)
	if err != nil {
		return httpapi.CombinedView{}, err
	}
	text, err := w.service.Combine(ctx, actor, id, file, field)
	if err != nil {
		return httpapi.CombinedView{}, err
	}
	return httpapi.CombinedView{Text: text}, nil
}
