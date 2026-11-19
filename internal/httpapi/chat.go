package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ChatView struct {
	Available bool `json:"available"`
}

type ChatService interface {
	Chat(ctx context.Context, actor Actor, organisation, project string) (ChatView, error)
	Ask(ctx context.Context, actor Actor, organisation, project, text string, about *ChatSelection) error
	Listen(ctx context.Context, actor Actor, organisation, project string, emit func(any) error) error
	Stop(ctx context.Context, actor Actor, organisation, project string) error
	Where(ctx context.Context, actor Actor, organisation, project string, about ChatSelection) (WhereView, error)
}

type ChatSelection struct {
	Text   string `json:"text"`
	Before string `json:"before"`
	After  string `json:"after"`
	Path   string `json:"path"`
}

type askRequest struct {
	Text      string         `json:"text"`
	Selection *ChatSelection `json:"selection"`
}

const (
	mostSelected = 2000
	mostAround   = 200
	mostPath     = 1000
)

func lastRunes(text string, most int) string {
	runes := []rune(text)
	if len(runes) <= most {
		return text
	}
	return string(runes[len(runes)-most:])
}

func firstRunes(text string, most int) string {
	runes := []rune(text)
	if len(runes) <= most {
		return text
	}
	return string(runes[:most])
}

func (s *ChatSelection) bounded() *ChatSelection {
	if s == nil || s.Text == "" {
		return nil
	}
	return &ChatSelection{
		Text:   firstRunes(s.Text, mostSelected),
		Before: lastRunes(s.Before, mostAround),
		After:  firstRunes(s.After, mostAround),
		Path:   firstRunes(s.Path, mostPath),
	}
}

type WhereView struct {
	Kind      string `json:"kind"`
	Page      string `json:"page,omitempty"`
	PageTitle string `json:"pageTitle,omitempty"`
	Field     string `json:"field,omitempty"`
}

func chatHandlers(mux *http.ServeMux, service ChatService) {
	base := "/api/organisations/{organisation}/projects/{project}/chat"

	mux.Handle("GET "+base, RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, err := service.Chat(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})))

	mux.Handle("POST "+base, RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body askRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}
		if err := service.Ask(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"), r.PathValue("project"), body.Text, body.Selection.bounded()); err != nil {
			writeTenancyError(w, err)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})))
	mux.HandleFunc(base, MethodNotAllowed)

	mux.Handle("POST "+base+"/stop", RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := service.Stop(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"), r.PathValue("project")); err != nil {
			writeTenancyError(w, err)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})))
	mux.HandleFunc(base+"/stop", MethodNotAllowed)

	mux.Handle("POST "+base+"/where", RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body ChatSelection
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}
		view, err := service.Where(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"), r.PathValue("project"), *body.bounded())
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})))
	mux.HandleFunc(base+"/where", MethodNotAllowed)

	mux.Handle("GET "+base+"/events", RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controller := http.NewResponseController(w)
		started := false
		err := service.Listen(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"), r.PathValue("project"), func(event any) error {
			if !started {
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("X-Accel-Buffering", "no")
				w.WriteHeader(http.StatusOK)
				started = true
			}
			data, err := json.Marshal(event)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				return err
			}
			return controller.Flush()
		})
		if err != nil && !started {
			writeTenancyError(w, err)
		}
	})))
	mux.HandleFunc(base+"/events", MethodNotAllowed)
}
