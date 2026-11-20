package httpapi

import (
	"context"
	"net/http"
)

type WorkspaceService interface {
	OpenWorkspace(ctx context.Context, actor Actor, organisation, project string) (WorkspaceView, error)
	WorkspaceStanding(ctx context.Context, actor Actor, organisation, project string) (WorkspaceView, error)
	RestartWorkspace(ctx context.Context, actor Actor, organisation, project string) (WorkspaceView, error)
}

type WorkspaceStepView struct {
	Step  string `json:"step"`
	State string `json:"state"`
}

type WorkspaceView struct {
	State    string              `json:"state"`
	Address  string              `json:"address,omitempty"`
	Type     string              `json:"type,omitempty"`
	Open     string              `json:"open,omitempty"`
	OwnPages bool                `json:"ownPages,omitempty"`
	Reason   string              `json:"reason,omitempty"`
	Steps    []WorkspaceStepView `json:"steps,omitempty"`
	Log      []string            `json:"log,omitempty"`
	Backend  string              `json:"backend,omitempty"`
	Since    string              `json:"since,omitempty"`
	Entry    string              `json:"entry,omitempty"`
}

func openWorkspaceHandler(service WorkspaceService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, err := service.OpenWorkspace(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})
}

func restartWorkspaceHandler(service WorkspaceService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, err := service.RestartWorkspace(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})
}

func workspaceStandingHandler(service WorkspaceService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, err := service.WorkspaceStanding(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})
}
