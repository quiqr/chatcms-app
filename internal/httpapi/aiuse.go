package httpapi

import (
	"context"
	"net/http"
)

type AIUseView struct {
	RequestsPerMinute *int   `json:"requestsPerMinute"`
	TokensPerMonth    *int64 `json:"tokensPerMonth"`
	RequestsThisMonth int64  `json:"requestsThisMonth"`
	TokensThisMonth   int64  `json:"tokensThisMonth"`
}

type AIUseService interface {
	AIUse(ctx context.Context, who Actor, slug string) (AIUseView, error)
	SetAILimits(ctx context.Context, who Actor, slug string, requests *int, tokens *int64) (AIUseView, error)
}

type aiLimitsRequest struct {
	RequestsPerMinute *int   `json:"requestsPerMinute"`
	TokensPerMonth    *int64 `json:"tokensPerMonth"`
}

func aiUseHandlers(mux *http.ServeMux, service AIUseService) {
	path := "/api/administration/organisations/{organisation}/ai"
	mux.Handle("GET "+path, RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, err := service.AIUse(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})))
	mux.Handle("PUT "+path, RequireActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body aiLimitsRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}
		view, err := service.SetAILimits(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"), body.RequestsPerMinute, body.TokensPerMonth)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})))
	mux.HandleFunc(path, MethodNotAllowed)
}
