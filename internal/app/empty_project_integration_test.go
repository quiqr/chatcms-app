//go:build integration

package app_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAProjectWithNothingBroughtInSaysSo(t *testing.T) {
	server, _ := startServer(t, testConfig(t))

	boss, err := server.Auth().People().Create(t.Context(), "admin@technative.example", "The Admin", "a long enough passphrase")
	if err != nil {
		t.Fatal(err)
	}
	if err := server.MakeAdministrator(t.Context(), boss.ID); err != nil {
		t.Fatal(err)
	}

	admin := newClient(t, server.Addr())
	admin.warmUp()
	for _, step := range []struct {
		path string
		body map[string]any
		want int
	}{
		{"/api/sign-in", map[string]any{"email": "admin@technative.example", "password": "a long enough passphrase"}, http.StatusOK},
		{"/api/organisations", map[string]any{"slug": "technative", "name": "TechNative"}, http.StatusCreated},
		{"/api/organisations/technative/projects", map[string]any{"slug": "xmin", "name": "Xmin", "domain": ""}, http.StatusCreated},
	} {
		response := admin.do(http.MethodPost, step.path, step.body, true)
		response.Body.Close()
		if response.StatusCode != step.want {
			t.Fatalf("POST %s = %d, want %d", step.path, response.StatusCode, step.want)
		}
	}

	for _, path := range []string{"content", "images", "review", "standing", "versions"} {
		response := admin.do(http.MethodGet, "/api/organisations/technative/projects/xmin/"+path, nil, false)
		var said struct {
			Error struct {
				Key string `json:"key"`
			} `json:"error"`
		}
		_ = json.NewDecoder(response.Body).Decode(&said)
		response.Body.Close()
		if response.StatusCode != http.StatusConflict || said.Error.Key != "refused.nothing-brought-in-yet" {
			t.Errorf("GET %s = %d %q, want 409 refused.nothing-brought-in-yet", path, response.StatusCode, said.Error.Key)
		}
	}
}
