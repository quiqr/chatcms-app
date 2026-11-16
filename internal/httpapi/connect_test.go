package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
)

type stubConnect struct {
	asked string
	err   error
}

func (s *stubConnect) ConnectSite(_ context.Context, _ Actor, organisation, project, remote string) (ImportView, error) {
	s.asked = organisation + "/" + project + "=" + remote
	if s.err != nil {
		return ImportView{}, s.err
	}
	return ImportView{Kind: "template", Remote: remote, Ready: true}, nil
}

func withConnect(service ConnectService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "implementer@example.org"}
	}
	return New(Options{
		Logger:  slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health:  func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:    &stubAuth{known: known},
		Connect: service,
	})
}

const connectPath = "/api/organisations/technative/projects/bakery/remote"

func TestConnectingNeedsAPerson(t *testing.T) {
	recorder := call(t, withConnect(&stubConnect{}, false), http.MethodPut, connectPath, map[string]string{"remote": "x"})
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", recorder.Code)
	}
}

func TestConnectingASite(t *testing.T) {
	service := &stubConnect{}
	recorder := signedInCall(t, withConnect(service, true), http.MethodPut, connectPath,
		map[string]string{"remote": "git@example.org:bakery.git"})

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	var view ImportView
	if err := json.Unmarshal(recorder.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	if view.Remote != "git@example.org:bakery.git" || view.Hosted {
		t.Errorf("view = %+v, want the new remote and not kept", view)
	}
	if service.asked != "technative/bakery=git@example.org:bakery.git" {
		t.Errorf("asked = %q", service.asked)
	}
}

func TestConnectingRefused(t *testing.T) {
	service := &stubConnect{err: &RefusedError{Key: "refused.connect-not-empty", Reason: "that repository is not empty"}}
	recorder := signedInCall(t, withConnect(service, true), http.MethodPut, connectPath, map[string]string{"remote": "x"})

	if recorder.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "refused.connect-not-empty") {
		t.Errorf("body = %s, want the key", recorder.Body.String())
	}
}

func TestConnectingOnlyByPut(t *testing.T) {
	recorder := signedInCall(t, withConnect(&stubConnect{}, true), http.MethodGet, connectPath, nil)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", recorder.Code)
	}
}
