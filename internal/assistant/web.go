package assistant

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/policy"
	"github.com/quiqr/chatcms-app/internal/tenancy"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

type Locator interface {
	Locate(ctx context.Context, key workspace.Key, path string) (string, error)
}

type Web struct {
	chat    *Chat
	tenancy *tenancy.Store
	under   func(organisation, project string) string
	locator Locator
	logger  *slog.Logger
}

func (w *Web) Locating(locator Locator, logger *slog.Logger) {
	w.locator = locator
	w.logger = logger
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

func (w *Web) Ask(ctx context.Context, who httpapi.Actor, organisation, project, text string, about *httpapi.ChatSelection) error {
	key, err := w.key(ctx, who, organisation, project)
	if err != nil {
		return err
	}
	if about != nil && strings.TrimSpace(text) != "" {
		page := ""
		if w.locator != nil {
			var err error
			if page, err = w.locator.Locate(ctx, key, about.Path); err != nil {
				w.logger.WarnContext(ctx, "the page of a selection could not be found", "path", about.Path, "error", err)
			}
		}
		text = AboutASelection(text, *about, page)
	}
	return refusal(w.chat.Send(ctx, key, w.under(organisation, project), text))
}

func AboutASelection(text string, about httpapi.ChatSelection, page string) string {
	var b strings.Builder
	b.WriteString(strings.TrimSpace(text))
	fmt.Fprintf(&b, "\n\nThe person selected this on the preview page %s:\n%q\n", about.Path, about.Text)
	if about.Before != "" {
		fmt.Fprintf(&b, "Just before it: %q\n", about.Before)
	}
	if about.After != "" {
		fmt.Fprintf(&b, "Just after it: %q\n", about.After)
	}
	if page != "" {
		fmt.Fprintf(&b, "It is on the page with key %q. Read it with read_page, change it with update_page as asked, and change nothing else.", page)
	} else {
		b.WriteString("No page was found for this address. Find the page that holds this text with list_pages and read_page; if no page holds it, say it comes from the site's design, which you cannot change.")
	}
	return b.String()
}

func (w *Web) Listen(ctx context.Context, who httpapi.Actor, organisation, project string, emit func(any) error) error {
	key, err := w.key(ctx, who, organisation, project)
	if err != nil {
		return err
	}
	return refusal(w.chat.Events(ctx, key, w.under(organisation, project), func(event Event) error { return emit(event) }))
}

var _ httpapi.ChatService = (*Web)(nil)
