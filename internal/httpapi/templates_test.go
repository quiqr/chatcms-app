package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"testing"
)

type stubTemplates struct {
	view  TemplatesView
	shot  []byte
	kind  string
	err   error
	asked string
}

func (s *stubTemplates) ListTemplates(_ context.Context, _ Actor, organisation string) (TemplatesView, error) {
	s.asked = organisation
	return s.view, s.err
}

func (s *stubTemplates) TemplateScreenshot(_ context.Context, _ Actor, organisation, template string) ([]byte, string, error) {
	s.asked = organisation + "/" + template
	return s.shot, s.kind, s.err
}

func (s *stubTemplates) ReadHugo(_ context.Context, _ Actor, organisation, project string) (ProjectHugoView, error) {
	s.asked = organisation + "/" + project
	return ProjectHugoView{Version: "0.101.0", Installed: s.view.Hugo}, s.err
}

func (s *stubTemplates) SetHugo(_ context.Context, _ Actor, organisation, project, version string, extended bool) (ProjectHugoView, error) {
	s.asked = organisation + "/" + project + "=" + version
	return ProjectHugoView{Version: version, Extended: extended}, s.err
}

func withTemplates(service TemplateService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "implementer@example.org"}
	}
	return New(Options{
		Logger:    slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health:    func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:      &stubAuth{known: known},
		Templates: service,
	})
}

const templatesBase = "/api/organisations/technative/templates"

func TestTemplatesNeedAPerson(t *testing.T) {
	handler := withTemplates(&stubTemplates{}, false)

	for _, path := range []string{templatesBase, templatesBase + "/Summer/screenshot"} {
		if recorder := call(t, handler, http.MethodGet, path, nil); recorder.Code != http.StatusUnauthorized {
			t.Errorf("GET %s = %d, want 401", path, recorder.Code)
		}
	}
}

func TestListingTemplates(t *testing.T) {
	service := &stubTemplates{view: TemplatesView{
		Offered: true, Available: true,
		Templates: []TemplateView{{ID: "Summer", Name: "Summer", Model: []string{"base"}}},
		Hugo:      HugoView{Version: "0.139.0", Extended: true},
	}}
	recorder := signedInCall(t, withTemplates(service, true), http.MethodGet, templatesBase, nil)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	var view TemplatesView
	if err := json.Unmarshal(recorder.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	if !view.Available || len(view.Templates) != 1 || view.Hugo.Version != "0.139.0" || service.asked != "technative" {
		t.Errorf("view = %+v, asked %q", view, service.asked)
	}
}

func TestTemplatesRefusedOrMissing(t *testing.T) {
	for err, want := range map[error]int{ErrNotAllowed: http.StatusForbidden, ErrMissing: http.StatusNotFound} {
		handler := withTemplates(&stubTemplates{err: err}, true)
		if recorder := signedInCall(t, handler, http.MethodGet, templatesBase, nil); recorder.Code != want {
			t.Errorf("%v gave %d, want %d", err, recorder.Code, want)
		}
		shot := signedInCall(t, handler, http.MethodGet, templatesBase+"/Summer/screenshot", nil)
		if shot.Code != want {
			t.Errorf("screenshot with %v gave %d, want %d", err, shot.Code, want)
		}
	}
}

func TestAScreenshotIsServedAsAPicture(t *testing.T) {
	service := &stubTemplates{shot: []byte("\x89PNG\r\n\x1a\n...."), kind: "image/png"}
	recorder := signedInCall(t, withTemplates(service, true), http.MethodGet, templatesBase+"/Summer/screenshot", nil)

	if recorder.Code != http.StatusOK || recorder.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("status %d, type %q", recorder.Code, recorder.Header().Get("Content-Type"))
	}
	if recorder.Header().Get("X-Content-Type-Options") != "nosniff" || recorder.Header().Get("Cache-Control") == "" {
		t.Errorf("headers = %v", recorder.Header())
	}
	if service.asked != "technative/Summer" || recorder.Body.Len() != len(service.shot) {
		t.Errorf("asked %q, %d bytes", service.asked, recorder.Body.Len())
	}
}

func TestTemplatesOnlyAnswerGet(t *testing.T) {
	handler := withTemplates(&stubTemplates{}, true)
	if recorder := signedInCall(t, handler, http.MethodPost, templatesBase, nil); recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST = %d, want 405", recorder.Code)
	}
}

func TestReadingAndSettingAProjectsHugo(t *testing.T) {
	service := &stubTemplates{}
	handler := withTemplates(service, true)
	path := "/api/organisations/technative/projects/korenaar/hugo"

	if recorder := signedInCall(t, handler, http.MethodGet, path, nil); recorder.Code != http.StatusOK {
		t.Errorf("GET = %d", recorder.Code)
	}
	recorder := signedInCall(t, handler, http.MethodPut, path, map[string]any{"version": "0.139.0", "extended": true})
	if recorder.Code != http.StatusOK || service.asked != "technative/korenaar=0.139.0" {
		t.Errorf("PUT = %d, asked %q", recorder.Code, service.asked)
	}
	if recorder := call(t, handler, http.MethodGet, path, nil); recorder.Code != http.StatusUnauthorized {
		t.Errorf("anonymous GET = %d, want 401", recorder.Code)
	}
}
