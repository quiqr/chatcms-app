package httpapi

import (
	"context"
	"net/http"
)

type TemplateService interface {
	ListTemplates(ctx context.Context, actor Actor, organisation string) (TemplatesView, error)
	TemplateScreenshot(ctx context.Context, actor Actor, organisation, template string) ([]byte, string, error)
	ReadHugo(ctx context.Context, actor Actor, organisation, project string) (ProjectHugoView, error)
	SetHugo(ctx context.Context, actor Actor, organisation, project, version string, extended bool) (ProjectHugoView, error)
}

type ProjectHugoView struct {
	Version   string   `json:"version"`
	Extended  bool     `json:"extended"`
	Installed HugoView `json:"installed"`
}

type hugoRequest struct {
	Version  string `json:"version"`
	Extended bool   `json:"extended"`
}

type TemplatesView struct {
	Offered   bool           `json:"offered"`
	Available bool           `json:"available"`
	Templates []TemplateView `json:"templates"`
	Hugo      HugoView       `json:"hugo"`
}

type HugoView struct {
	Version  string `json:"version"`
	Extended bool   `json:"extended"`
}

type TemplateView struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Author        string   `json:"author"`
	AuthorURL     string   `json:"authorUrl"`
	License       string   `json:"license"`
	LicenseURL    string   `json:"licenseUrl"`
	Homepage      string   `json:"homepage"`
	DemoURL       string   `json:"demoUrl"`
	HugoTheme     string   `json:"hugoTheme"`
	HugoVersion   string   `json:"hugoVersion"`
	Extended      bool     `json:"extended"`
	Model         []string `json:"model"`
	Forms         int      `json:"forms"`
	SourceURL     string   `json:"sourceUrl"`
	ScreenshotURL string   `json:"screenshotUrl"`
}

func templateHandlers(mux *http.ServeMux, service TemplateService) {
	base := "/api/organisations/{organisation}/templates"

	mux.Handle("GET "+base, RequireActor(listTemplatesHandler(service)))
	mux.HandleFunc(base, MethodNotAllowed)

	mux.Handle("GET "+base+"/{template}/screenshot", RequireActor(templateScreenshotHandler(service)))
	mux.HandleFunc(base+"/{template}/screenshot", MethodNotAllowed)

	hugo := "/api/organisations/{organisation}/projects/{project}/hugo"
	mux.Handle("GET "+hugo, RequireActor(readHugoHandler(service)))
	mux.Handle("PUT "+hugo, RequireActor(setHugoHandler(service)))
	mux.HandleFunc(hugo, MethodNotAllowed)
}

func readHugoHandler(service TemplateService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, err := service.ReadHugo(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})
}

func setHugoHandler(service TemplateService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body hugoRequest
		if err := Decode(w, r, &body); err != nil {
			WriteDecodeError(w, err)
			return
		}
		view, err := service.SetHugo(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("project"), body.Version, body.Extended)
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})
}

func listTemplatesHandler(service TemplateService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		view, err := service.ListTemplates(r.Context(), ActorFrom(r.Context()), r.PathValue("organisation"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, view)
	})
}

func templateScreenshotHandler(service TemplateService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, kind, err := service.TemplateScreenshot(r.Context(), ActorFrom(r.Context()),
			r.PathValue("organisation"), r.PathValue("template"))
		if err != nil {
			writeTenancyError(w, err)
			return
		}
		w.Header().Set("Content-Type", kind)
		w.Header().Set("Cache-Control", "private, max-age=86400")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})
}
