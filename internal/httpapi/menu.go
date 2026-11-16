package httpapi

import (
	"context"
	"net/http"
	"strconv"
)

type MenuService interface {
	ContentMenu(ctx context.Context, actor Actor, organisation, project string) (ContentMenuView, error)
	CollectionItems(ctx context.Context, actor Actor, organisation, project, collection string, query ItemsQuery) (CollectionItemsView, error)
	CreateItem(ctx context.Context, actor Actor, organisation, project, collection string, asked NewItemRequest) (PageView, error)
	ItemAddress(ctx context.Context, actor Actor, organisation, project, collection string, asked NewItemRequest) (AddressView, error)
	Search(ctx context.Context, actor Actor, organisation, project, q string) (SearchView, error)
}

type NewItemRequest struct {
	Title  string `json:"title"`
	Folder string `json:"folder"`
	Slug   string `json:"slug"`
}

type AddressView struct {
	Slug    string `json:"slug"`
	Folder  string `json:"folder,omitempty"`
	Address string `json:"address"`
	Free    bool   `json:"free"`
	File    string `json:"file,omitempty"`
}

type SearchView struct {
	Groups []SearchGroupView `json:"groups"`
}

type SearchGroupView struct {
	Title string           `json:"title"`
	Items []SearchItemView `json:"items"`
}

type SearchItemView struct {
	Key   string `json:"key"`
	Title string `json:"title"`
	Kind  string `json:"kind"`
}

type ItemsQuery struct {
	Q        string
	Folder   string
	Sort     string
	Page     int
	PageSize int
	Around   string
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
	Key        string                 `json:"key"`
	Title      string                 `json:"title"`
	ItemName   string                 `json:"itemName"`
	Items      []CollectionItemView   `json:"items"`
	Total      int                    `json:"total"`
	All        int                    `json:"all"`
	Sort       string                 `json:"sort"`
	Folders    []CollectionFolderView `json:"folders"`
	ShowsPaths bool                   `json:"showsPaths"`
	HasImages  bool                   `json:"hasImages"`
	Around     *AroundView            `json:"around,omitempty"`
}

type AroundView struct {
	Index    int    `json:"index"`
	Previous string `json:"previous,omitempty"`
	Next     string `json:"next,omitempty"`
}

type CollectionItemView struct {
	Key       string `json:"key"`
	Title     string `json:"title"`
	Folder    string `json:"folder,omitempty"`
	Status    string `json:"status"`
	By        string `json:"by,omitempty"`
	Date      string `json:"date,omitempty"`
	Author    string `json:"author,omitempty"`
	Summary   string `json:"summary,omitempty"`
	Thumbnail string `json:"thumbnail,omitempty"`
	Edited    string `json:"edited,omitempty"`
	Path      string `json:"path,omitempty"`
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
		asked := r.URL.Query()
		view, err := service.CollectionItems(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), r.PathValue("collection"), ItemsQuery{
				Q:        asked.Get("q"),
				Folder:   asked.Get("folder"),
				Sort:     asked.Get("sort"),
				Page:     number(asked.Get("page")),
				PageSize: number(asked.Get("pageSize")),
				Around:   asked.Get("around"),
			})
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})))
	mux.Handle("POST "+base+"/collections/{collection}/items", RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body NewItemRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}
		page, err := service.CreateItem(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), r.PathValue("collection"), body)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusCreated, page)
	})))
	mux.HandleFunc(base+"/collections/{collection}/items", MethodNotAllowed)

	mux.Handle("GET "+base+"/collections/{collection}/address", RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asked := r.URL.Query()
		view, err := service.ItemAddress(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), r.PathValue("collection"), NewItemRequest{
				Title: asked.Get("title"), Folder: asked.Get("folder"), Slug: asked.Get("slug"),
			})
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})))
	mux.HandleFunc(base+"/collections/{collection}/address", MethodNotAllowed)

	mux.Handle("GET "+base+"/search", RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, err := service.Search(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), r.URL.Query().Get("q"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})))
	mux.HandleFunc(base+"/search", MethodNotAllowed)
}

func number(said string) int {
	value, err := strconv.Atoi(said)
	if err != nil || value < 0 {
		return 0
	}
	return value
}
