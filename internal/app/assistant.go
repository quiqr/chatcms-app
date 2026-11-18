package app

import (
	"context"
	"net/http"

	"github.com/quiqr/chatcms-app/internal/aigateway"
	"github.com/quiqr/chatcms-app/internal/assistant"
	"github.com/quiqr/chatcms-app/internal/content"
	"github.com/quiqr/chatcms-app/internal/gitsync"
	"github.com/quiqr/chatcms-app/internal/policy"
	"github.com/quiqr/chatcms-app/internal/tenancy"
)

type assistantPages struct {
	content *content.Service
	people  *tenancy.Store
	sites   *gitsync.Sites
}

func (p assistantPages) actor(ctx context.Context, caller aigateway.Caller) (policy.Actor, error) {
	return p.people.ActorFor(ctx, caller.Person)
}

func (p assistantPages) List(ctx context.Context, caller aigateway.Caller) ([]assistant.PageRef, error) {
	actor, err := p.actor(ctx, caller)
	if err != nil {
		return nil, err
	}
	sections, err := p.content.List(ctx, actor, caller.Project)
	if err != nil {
		return nil, err
	}
	pages := []assistant.PageRef{}
	for _, section := range sections {
		for _, item := range section.Items {
			pages = append(pages, assistant.PageRef{Key: item.Key, Title: item.Title, Section: section.Title})
		}
	}
	return pages, nil
}

func fieldsOf(controls []content.Control, into *[]assistant.Field) {
	for _, control := range controls {
		if len(control.Controls) > 0 && control.Value == nil {
			fieldsOf(control.Controls, into)
			continue
		}
		*into = append(*into, assistant.Field{Path: control.Path, Label: control.Title, Value: control.Value})
	}
}

func (p assistantPages) Read(ctx context.Context, caller aigateway.Caller, key string) (assistant.Page, error) {
	actor, err := p.actor(ctx, caller)
	if err != nil {
		return assistant.Page{}, err
	}
	editable, err := p.content.Read(ctx, actor, caller.Project, key)
	if err != nil {
		return assistant.Page{}, err
	}
	page := assistant.Page{Key: editable.Key, Title: editable.Title, Fields: []assistant.Field{}}
	fieldsOf(editable.Controls, &page.Fields)
	if editable.HasBody {
		text := editable.Body
		page.Text = &text
	}
	return page, nil
}

func (p assistantPages) latest(ctx context.Context, caller aigateway.Caller) string {
	repository, err := openFor(ctx, p.sites, caller.Person.String(), caller.Project)
	if err != nil {
		return ""
	}
	defer repository.Close()
	versions, err := repository.History(ctx, 1)
	if err != nil || len(versions) == 0 {
		return ""
	}
	return versions[0].ID
}

func (p assistantPages) Update(ctx context.Context, caller aigateway.Caller, key string, fields map[string]any, text *string) (assistant.Change, error) {
	actor, err := p.actor(ctx, caller)
	if err != nil {
		return assistant.Change{}, err
	}
	before := p.latest(ctx, caller)
	saved, err := p.content.Save(ctx, actor, caller.Project, key, content.Edit{Values: fields, Body: text})
	if err != nil {
		return assistant.Change{}, err
	}
	return assistant.Change{Page: saved.Key, Title: saved.Title, Before: before, After: p.latest(ctx, caller)}, nil
}

var _ assistant.Pages = assistantPages{}

func (s *Server) startAssistant(service *content.Service, tenancyStore *tenancy.Store) {
	if s.aiMux == nil || s.workspaces == nil {
		return
	}
	s.chat = assistant.NewChat(s.workspaces)
	tools := assistant.NewTools(s.aiTokens, assistantPages{content: service, people: tenancyStore, sites: s.sites}, s.chat)
	environment := assistant.NewAgentEnvironment(s.aiTokens, "/run/agent")
	s.aiMux.Handle("/mcp", s.fromAWorkspace(tools))
	s.aiMux.Handle("/agent/env", s.fromAWorkspace(environment))
}

func (s *Server) fromAWorkspace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.aiGateway.FromAWorkspace(r) {
			http.Error(w, "this answers workspaces only", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
