package httpapi

import (
	"context"
	"io"
	"mime"
	"net/http"
	"time"
)

type FilesService interface {
	ProjectFiles(ctx context.Context, actor Actor, organisation, project string) (FilesView, error)
	ProjectFile(ctx context.Context, actor Actor, organisation, project, path string) (FileView, error)
	SaveProjectFile(ctx context.Context, actor Actor, organisation, project, path string, save FileSave) (FileView, error)
	OpenProjectFile(ctx context.Context, actor Actor, organisation, project, path string) (FileContent, error)
}

type FileContent struct {
	Name     string
	Preview  string
	Type     string
	Modified time.Time
	Content  io.ReadSeeker
	Close    func() error
}

type FileEntryView struct {
	Path string `json:"path"`
	Dir  bool   `json:"dir,omitempty"`
	Link bool   `json:"link,omitempty"`
	Size int64  `json:"size,omitempty"`
}

type FilesView struct {
	Entries []FileEntryView `json:"entries"`
}

type FileView struct {
	Path     string `json:"path"`
	Text     string `json:"text"`
	Sum      string `json:"sum,omitempty"`
	Unusable string `json:"unusable,omitempty"`
	Preview  string `json:"preview,omitempty"`
}

type FileSave struct {
	Text string `json:"text"`
	Sum  string `json:"sum"`
}

func filesHandlers(mux *http.ServeMux, service FilesService) {
	base := "/api/organisations/{organisation}/projects/{project}/files"
	mux.Handle("GET "+base, RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, err := service.ProjectFiles(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})))
	mux.HandleFunc(base, MethodNotAllowed)
	mux.Handle("GET "+base+"/{path...}", RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, err := service.ProjectFile(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"), r.PathValue("project"), r.PathValue("path"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})))
	mux.Handle("PUT "+base+"/{path...}", RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body FileSave
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}
		view, err := service.SaveProjectFile(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"), r.PathValue("project"), r.PathValue("path"), body)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})))
	mux.HandleFunc(base+"/{path...}", MethodNotAllowed)

	content := "/api/organisations/{organisation}/projects/{project}/file-content/{path...}"
	mux.Handle("GET "+content, RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, err := service.OpenProjectFile(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"), r.PathValue("project"), r.PathValue("path"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		defer file.Close()
		w.Header().Set("Content-Type", file.Type)
		w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": file.Name}))
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "private, max-age=0")
		if file.Preview != "pdf" {
			w.Header().Set("Content-Security-Policy", "sandbox")
		}
		http.ServeContent(w, r, file.Name, file.Modified, file.Content)
	})))
	mux.HandleFunc(content, MethodNotAllowed)
}
