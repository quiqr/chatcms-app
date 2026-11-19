package httpapi

import (
	"context"
	"io"
	"mime"
	"net/http"
	"path"
	"strconv"
	"time"
)

type ArtifactView struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"createdAt"`
	PublishID string `json:"publishId,omitempty"`
}

type ArtifactService interface {
	Artifacts(ctx context.Context, actor Actor, organisation, project string) ([]ArtifactView, error)
	OpenArtifact(ctx context.Context, actor Actor, organisation, project, id string) (ArtifactView, io.ReadSeekCloser, time.Time, error)
}

func artifactHandlers(mux *http.ServeMux, service ArtifactService) {
	base := "/api/organisations/{organisation}/projects/{project}/artifacts"

	mux.Handle("GET "+base, RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		views, err := service.Artifacts(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, views)
	})))
	mux.HandleFunc(base, MethodNotAllowed)

	mux.Handle("GET "+base+"/{id}", RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, file, when, err := service.OpenArtifact(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"), r.PathValue("project"), r.PathValue("id"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		defer file.Close()
		name := path.Base(view.Name)
		kind := mime.TypeByExtension(path.Ext(name))
		if kind == "" {
			kind = "application/octet-stream"
		}
		w.Header().Set("Content-Type", kind)
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "sandbox")
		w.Header().Set("Cache-Control", "private, max-age=0")
		w.Header().Set("X-Artifact-Size", strconv.FormatInt(view.Size, 10))
		http.ServeContent(w, r, name, when, file)
	})))
	mux.HandleFunc(base+"/{id}", MethodNotAllowed)
}
