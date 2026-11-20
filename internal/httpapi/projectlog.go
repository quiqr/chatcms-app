package httpapi

import (
	"context"
	"net/http"
)

type ProjectLogService interface {
	ProjectLog(ctx context.Context, actor Actor, organisation, project, kind, before string) (LogView, error)
	LogEntry(ctx context.Context, actor Actor, organisation, project, entry string) (LogDetailView, error)
}

type LogEntryView struct {
	Entry    string `json:"entry"`
	Kind     string `json:"kind"`
	At       string `json:"at"`
	Finished string `json:"finished,omitempty"`
	State    string `json:"state"`
	What     string `json:"what,omitempty"`
	Who      string `json:"who,omitempty"`
	Millis   *int   `json:"millis,omitempty"`
}

type LogView struct {
	Entries []LogEntryView `json:"entries"`
	More    bool           `json:"more"`
}

type HugoFactsView struct {
	Asked   string `json:"asked"`
	Program string `json:"program"`
	Because string `json:"because"`
	Printed string `json:"printed"`
}

type LogStepView struct {
	Step   string `json:"step"`
	State  string `json:"state"`
	Detail string `json:"detail,omitempty"`
}

type LogDetailView struct {
	LogEntryView
	Hugo   *HugoFactsView `json:"hugo,omitempty"`
	Steps  []LogStepView  `json:"steps"`
	Log    string         `json:"log,omitempty"`
	Reason string         `json:"reason,omitempty"`
	Exit   *int           `json:"exit,omitempty"`
}

func projectLogHandlers(mux *http.ServeMux, service ProjectLogService) {
	base := "/api/organisations/{organisation}/projects/{project}/log"
	mux.Handle("GET "+base, RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, err := service.ProjectLog(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"), r.PathValue("project"),
			r.URL.Query().Get("kind"), r.URL.Query().Get("before"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})))
	mux.Handle("GET "+base+"/{entry}", RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, err := service.LogEntry(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"), r.PathValue("project"), r.PathValue("entry"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})))
	mux.HandleFunc(base, MethodNotAllowed)
	mux.HandleFunc(base+"/{entry}", MethodNotAllowed)
}
