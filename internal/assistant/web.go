package assistant

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/policy"
	"github.com/quiqr/chatcms-app/internal/tenancy"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

type Web struct {
	chat    *Chat
	tenancy *tenancy.Store
	under   func(organisation, project string) string
}

func NewWeb(chat *Chat, tenancyStore *tenancy.Store, under func(organisation, project string) string) *Web {
	return &Web{chat: chat, tenancy: tenancyStore, under: under}
}

func (w *Web) key(ctx context.Context, who httpapi.Actor, organisation, project string) (workspace.Key, error) {
	if who.Anonymous || who.ID == "" {
		return workspace.Key{}, httpapi.ErrNotAllowed
	}
	person, err := uuid.Parse(who.ID)
	if err != nil {
		return workspace.Key{}, httpapi.ErrNotAllowed
	}
	actor, err := w.tenancy.ActorFor(ctx, person)
	if err != nil {
		return workspace.Key{}, err
	}
	found, err := w.tenancy.ProjectBySlug(ctx, organisation, project)
	if errors.Is(err, tenancy.ErrNotFound) {
		return workspace.Key{}, httpapi.ErrMissing
	}
	if err != nil {
		return workspace.Key{}, err
	}
	resource := policy.Project(found.OrganisationID.String(), found.ID.String())
	if !policy.Can(actor, policy.ActionReadContent, resource) {
		return workspace.Key{}, httpapi.ErrMissing
	}
	return workspace.Key{Person: person, Project: found.ID}, nil
}

var refusedUnavailable = &httpapi.RefusedError{
	Key:    "refused.assistant-not-available",
	Reason: "the assistant is not available on this installation",
}

var refusedNoAnswer = &httpapi.RefusedError{
	Key:    "refused.assistant-did-not-answer",
	Reason: "the assistant did not answer; try again in a moment",
}

func refusal(err error) error {
	switch {
	case errors.Is(err, ErrUnavailable):
		return refusedUnavailable
	case errors.Is(err, ErrNoAgent):
		return refusedNoAnswer
	default:
		return err
	}
}

func (w *Web) Chat(ctx context.Context, who httpapi.Actor, organisation, project string) (httpapi.ChatView, error) {
	if _, err := w.key(ctx, who, organisation, project); err != nil {
		return httpapi.ChatView{}, err
	}
	return httpapi.ChatView{Available: w.chat.Available()}, nil
}

func (w *Web) Ask(ctx context.Context, who httpapi.Actor, organisation, project, text string) error {
	key, err := w.key(ctx, who, organisation, project)
	if err != nil {
		return err
	}
	return refusal(w.chat.Send(ctx, key, w.under(organisation, project), text))
}

func (w *Web) Listen(ctx context.Context, who httpapi.Actor, organisation, project string, emit func(any) error) error {
	key, err := w.key(ctx, who, organisation, project)
	if err != nil {
		return err
	}
	return refusal(w.chat.Events(ctx, key, w.under(organisation, project), func(event Event) error { return emit(event) }))
}

var _ httpapi.ChatService = (*Web)(nil)
