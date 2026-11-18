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
	Ask(ctx context.Context, actor Actor, organisation, project, text string) error
	Listen(ctx context.Context, actor Actor, organisation, project string, emit func(any) error) error
}

type askRequest struct {
	Text string `json:"text"`
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
		if err := service.Ask(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"), r.PathValue("project"), body.Text); err != nil {
			writeTenancyError(w, err)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})))
	mux.HandleFunc(base, MethodNotAllowed)

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
