package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"testing"
)

type stubWorkspace struct {
	view   WorkspaceView
	err    error
	called []string
}

func (s *stubWorkspace) OpenWorkspace(_ context.Context, _ Actor, organisation, project string) (WorkspaceView, error) {
	s.called = append(s.called, "open "+organisation+"/"+project)
	return s.view, s.err
}

func (s *stubWorkspace) WorkspaceStanding(_ context.Context, _ Actor, organisation, project string) (WorkspaceView, error) {
	s.called = append(s.called, "standing "+organisation+"/"+project)
	return s.view, s.err
}

func withWorkspace(service WorkspaceService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "anna@example.org"}
	}
	return New(Options{
		Logger:    slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health:    func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:      &stubAuth{known: known},
		Workspace: service,
	})
}

const workspacePath = "/api/organisations/technative/projects/korenaar/workspace"

func TestWorkspaceEndpointsNeedAPerson(t *testing.T) {
	handler := withWorkspace(&stubWorkspace{}, false)
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		if recorder := call(t, handler, method, workspacePath, map[string]any{}); recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s = %d, want 401", method, recorder.Code)
		}
	}
}

func TestOpeningAWorkspaceSaysHowItStands(t *testing.T) {
	service := &stubWorkspace{view: WorkspaceView{
		State: "starting",
		Steps: []WorkspaceStepView{{Step: "starting-workspace", State: "running"}},
		Log:   []string{"fetching"},
	}}
	handler := withWorkspace(service, true)

	recorder := signedInCall(t, handler, http.MethodPost, workspacePath, nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("POST = %d: %s", recorder.Code, recorder.Body)
	}
	var view WorkspaceView
	if err := json.NewDecoder(recorder.Body).Decode(&view); err != nil {
		t.Fatal(err)
	}
	if view.State != "starting" || len(view.Steps) != 1 || view.Log[0] != "fetching" {
		t.Errorf("view = %+v", view)
	}

	if recorder := signedInCall(t, handler, http.MethodGet, workspacePath, nil); recorder.Code != http.StatusOK {
		t.Errorf("GET = %d", recorder.Code)
	}
	if len(service.called) != 2 || service.called[0] != "open technative/korenaar" || service.called[1] != "standing technative/korenaar" {
		t.Errorf("called = %v", service.called)
	}
}

func TestAWorkspaceRefusalIsPassedOn(t *testing.T) {
	handler := withWorkspace(&stubWorkspace{err: &RefusedError{Key: "refused.too-many-sites-being-previewed", Reason: "too many"}}, true)

	recorder := signedInCall(t, handler, http.MethodPost, workspacePath, nil)
	if recorder.Code < 400 || !json.Valid(recorder.Body.Bytes()) {
		t.Errorf("POST = %d %s, want the refusal", recorder.Code, recorder.Body)
	}
	if missing := withWorkspace(&stubWorkspace{err: ErrMissing}, true); signedInCall(t, missing, http.MethodGet, workspacePath, nil).Code != http.StatusNotFound {
		t.Errorf("a missing project is not a 404")
	}
}

func TestOtherMethodsOnAWorkspaceAreRefused(t *testing.T) {
	handler := withWorkspace(&stubWorkspace{}, true)
	if recorder := signedInCall(t, handler, http.MethodDelete, workspacePath, nil); recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("DELETE = %d, want 405", recorder.Code)
	}
}
