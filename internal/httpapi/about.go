package httpapi

import (
	"context"
	"net/http"
)

type RunsOnView struct {
	Hugo     string `json:"hugo"`
	Git      string `json:"git"`
	Postgres string `json:"postgres"`
	Go       string `json:"go"`
}

type LicenseView struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

type LinksView struct {
	Source       string `json:"source"`
	Changelog    string `json:"changelog"`
	Issues       string `json:"issues"`
	Contributors string `json:"contributors"`
}

type AboutView struct {
	Version  string      `json:"version"`
	Source   string      `json:"source"`
	Changed  bool        `json:"changed"`
	Describe string      `json:"describe"`
	RunsOn   RunsOnView  `json:"runsOn"`
	License  LicenseView `json:"license"`
	Links    LinksView   `json:"links"`
}

type AboutSource interface {
	About(ctx context.Context) AboutView
}

func aboutHandlers(mux *http.ServeMux, admin AdministrationService, about AboutSource) {
	mux.Handle("GET /api/administration/about", RequireActor(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if !admin.MayAdminister(r.Context(), ActorFrom(r.Context())) {
				writeTenancyError(w, ErrNotAllowed)
				return
			}
			WriteJSON(w, http.StatusOK, about.About(r.Context()))
		})))
	mux.HandleFunc("/api/administration/about", MethodNotAllowed)
}
