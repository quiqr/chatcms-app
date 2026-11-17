package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"testing"
)

type stubAbout struct{ asked int }

func (s *stubAbout) About(context.Context) AboutView {
	s.asked++
	return AboutView{
		Version:  "0.1.0",
		Source:   "a3bab43",
		Describe: "0.1.0 · a3bab43",
		RunsOn:   RunsOnView{Hugo: "0.152.2", Git: "2.51.0", Postgres: "16.10", Go: "go1.26.9"},
		License:  LicenseView{Name: "Apache 2.0", Text: "Apache License"},
		Links:    LinksView{Source: "https://github.com/quiqr/chatcms-app"},
	}
}

func withAbout(administers, signedIn bool, about AboutSource) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1"}
	}
	return New(Options{
		Logger:         slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Auth:           &stubAuth{known: known},
		Administration: &stubAdministration{administers: administers},
		About:          about,
	})
}

func TestAnAdministratorReadsAboutTheInstallation(t *testing.T) {
	about := &stubAbout{}
	recorder := signedInCall(t, withAbout(true, true, about), http.MethodGet, "/api/administration/about", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body)
	}

	var got AboutView
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Describe != "0.1.0 · a3bab43" || got.RunsOn.Postgres != "16.10" || got.License.Name != "Apache 2.0" ||
		got.Links.Source == "" {
		t.Errorf("answer = %+v", got)
	}
}

func TestSomebodyElseIsRefusedAboutTheInstallation(t *testing.T) {
	about := &stubAbout{}
	recorder := signedInCall(t, withAbout(false, true, about), http.MethodGet, "/api/administration/about", nil)
	if recorder.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", recorder.Code)
	}
	if about.asked != 0 {
		t.Error("the installation was asked about itself for somebody who may not know")
	}
}

func TestAboutNeedsSomebody(t *testing.T) {
	recorder := call(t, withAbout(true, false, &stubAbout{}), http.MethodGet, "/api/administration/about", nil)
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", recorder.Code)
	}
}

func TestAboutIsOnlyRead(t *testing.T) {
	recorder := signedInCall(t, withAbout(true, true, &stubAbout{}), http.MethodPost, "/api/administration/about", nil)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", recorder.Code)
	}
}
