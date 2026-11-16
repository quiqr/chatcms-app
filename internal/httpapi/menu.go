package httpapi

import (
	"context"
	"net/http"
)

type MenuService interface {
	ContentMenu(ctx context.Context, actor Actor, organisation, project string) (ContentMenuView, error)
	CollectionItems(ctx context.Context, actor Actor, organisation, project, collection string) (CollectionItemsView, error)
}

type ContentMenuView struct {
	Groups   []MenuGroupView `json:"groups"`
	Totals   MenuTotalsView  `json:"totals"`
	Fallback bool            `json:"fallback"`
	Warnings []string        `json:"warnings,omitempty"`
}

type MenuTotalsView struct {
	Singles     int `json:"singles"`
	Collections int `json:"collections"`
}

type MenuGroupView struct {
	Key   string          `json:"key"`
	Title string          `json:"title"`
	Items []MenuEntryView `json:"items"`
}

type MenuEntryView struct {
	Key        string `json:"key"`
	Kind       string `json:"kind"`
	Title      string `json:"title"`
	ItemName   string `json:"itemName,omitempty"`
	Count      *int   `json:"count,omitempty"`
	HasChanges bool   `json:"hasChanges"`
}

type CollectionItemsView struct {
	Key      string                 `json:"key"`
	Title    string                 `json:"title"`
	ItemName string                 `json:"itemName"`
	Items    []CollectionItemView   `json:"items"`
	Total    int                    `json:"total"`
	Folders  []CollectionFolderView `json:"folders"`
}

type CollectionItemView struct {
	Key    string `json:"key"`
	Title  string `json:"title"`
	Folder string `json:"folder,omitempty"`
	Status string `json:"status"`
	By     string `json:"by,omitempty"`
	Path   string `json:"path,omitempty"`
}

type CollectionFolderView struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

func menuHandlers(mux *http.ServeMux, service MenuService) {
	base := "/api/organisations/{organisation}/projects/{project}"

	mux.Handle("GET "+base+"/content-menu", RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, err := service.ContentMenu(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})))
	mux.HandleFunc(base+"/content-menu", MethodNotAllowed)

	mux.Handle("GET "+base+"/collections/{collection}/items", RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, err := service.CollectionItems(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), r.PathValue("collection"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})))
	mux.HandleFunc(base+"/collections/{collection}/items", MethodNotAllowed)
}
