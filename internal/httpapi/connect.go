package httpapi

import (
	"context"
	"net/http"
)

type ConnectService interface {
	ConnectSite(ctx context.Context, actor Actor, organisation, project, remote string) (ImportView, error)
}

type connectRequest struct {
	Remote string `json:"remote"`
}

func connectHandlers(mux *http.ServeMux, service ConnectService) {
	path := "/api/organisations/{organisation}/projects/{project}/remote"
	mux.Handle("PUT "+path, RequireActor(connectSiteHandler(service)))
	mux.HandleFunc(path, MethodNotAllowed)
}

func connectSiteHandler(service ConnectService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body connectRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}
		view, err := service.ConnectSite(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), body.Remote)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})
}
