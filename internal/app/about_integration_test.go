//go:build integration

package app_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAnAdministratorSeesWhatTheInstallationRunsOn(t *testing.T) {
	server, caller, password := signInHarness(t)
	person, err := server.Auth().People().Create(t.Context(), "boss@example.org", "Boss", password)
	if err != nil {
		t.Fatal(err)
	}
	if err := server.MakeAdministrator(t.Context(), person.ID); err != nil {
		t.Fatal(err)
	}

	signedIn := caller.do(http.MethodPost, "/api/sign-in",
		map[string]any{"email": "boss@example.org", "password": password}, true)
	signedIn.Body.Close()
	if signedIn.StatusCode != http.StatusOK {
		t.Fatalf("sign-in = %d", signedIn.StatusCode)
	}

	response := caller.do(http.MethodGet, "/api/administration/about", nil, false)
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("about = %d, want 200", response.StatusCode)
	}

	var got struct {
		Describe string `json:"describe"`
		RunsOn   struct {
			Git      string `json:"git"`
			Postgres string `json:"postgres"`
			Go       string `json:"go"`
		} `json:"runsOn"`
	}
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.RunsOn.Postgres == "" || got.RunsOn.Postgres[0] < '0' || got.RunsOn.Postgres[0] > '9' {
		t.Errorf("postgres = %q, want its version", got.RunsOn.Postgres)
	}
	if got.RunsOn.Git == "" {
		t.Error("git said no version")
	}
	if !strings.Contains(got.Describe, "development build") {
		t.Errorf("describe = %q, want a test build to say it is a development build", got.Describe)
	}
}

func TestSomebodyWhoDoesNotAdministerIsNotToldAbout(t *testing.T) {
	_, caller, password := signInHarness(t)
	signedIn := caller.do(http.MethodPost, "/api/sign-in",
		map[string]any{"email": "joris@example.org", "password": password}, true)
	signedIn.Body.Close()

	response := caller.do(http.MethodGet, "/api/administration/about", nil, false)
	response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Errorf("about = %d, want 403", response.StatusCode)
	}
}
