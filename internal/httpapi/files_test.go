package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"testing"
)

type stubFiles struct {
	asked []string
	err   error
}

func (s *stubFiles) ProjectFiles(_ context.Context, _ Actor, organisation, project string) (FilesView, error) {
	s.asked = append(s.asked, "tree "+organisation+"/"+project)
	return FilesView{Entries: []FileEntryView{{Path: "layouts", Dir: true}, {Path: "layouts/index.html", Size: 20}}}, s.err
}

func (s *stubFiles) ProjectFile(_ context.Context, _ Actor, organisation, project, path string) (FileView, error) {
	s.asked = append(s.asked, "read "+path)
	return FileView{Path: path, Text: "<h1>{{ .Title }}</h1>", Sum: "abc"}, s.err
}

func (s *stubFiles) SaveProjectFile(_ context.Context, _ Actor, organisation, project, path string, save FileSave) (FileView, error) {
	s.asked = append(s.asked, "save "+path+" "+save.Sum+" "+save.Text)
	return FileView{Path: path, Text: save.Text, Sum: "def"}, s.err
}

func withFiles(service FilesService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "anna@example.org"}
	}
	return New(Options{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health: func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:   &stubAuth{known: known},
		Files:  service,
	})
}

const filesPath = "/api/organisations/technative/projects/korenaar/files"

func TestFilesNeedAPerson(t *testing.T) {
	handler := withFiles(&stubFiles{}, false)
	for _, path := range []string{filesPath, filesPath + "/layouts/index.html"} {
		if recorder := call(t, handler, http.MethodGet, path, map[string]any{}); recorder.Code != http.StatusUnauthorized {
			t.Errorf("GET %s = %d, want 401", path, recorder.Code)
		}
	}
}

func TestFilesAreListedReadAndSaved(t *testing.T) {
	service := &stubFiles{}
	handler := withFiles(service, true)

	recorder := signedInCall(t, handler, http.MethodGet, filesPath, nil)
	var tree FilesView
	if err := json.NewDecoder(recorder.Body).Decode(&tree); err != nil || len(tree.Entries) != 2 || !tree.Entries[0].Dir {
		t.Fatalf("tree = %+v, %v (%d)", tree, err, recorder.Code)
	}

	recorder = signedInCall(t, handler, http.MethodGet, filesPath+"/layouts/partials/head.html", nil)
	var file FileView
	if err := json.NewDecoder(recorder.Body).Decode(&file); err != nil || file.Path != "layouts/partials/head.html" || file.Sum != "abc" {
		t.Fatalf("file = %+v, %v", file, err)
	}

	recorder = signedInCall(t, handler, http.MethodPut, filesPath+"/layouts/index.html", map[string]any{"text": "<h2/>", "sum": "abc"})
	if recorder.Code != http.StatusOK {
		t.Fatalf("PUT = %d: %s", recorder.Code, recorder.Body)
	}
	want := []string{"tree technative/korenaar", "read layouts/partials/head.html", "save layouts/index.html abc <h2/>"}
	for i, line := range want {
		if i >= len(service.asked) || service.asked[i] != line {
			t.Errorf("asked = %v, want %v", service.asked, want)
			break
		}
	}
}

func TestFilesAreRefusedWithoutTechnicalDetail(t *testing.T) {
	handler := withFiles(&stubFiles{err: ErrNotAllowed}, true)
	if recorder := signedInCall(t, handler, http.MethodGet, filesPath, nil); recorder.Code != http.StatusForbidden {
		t.Errorf("GET = %d, want 403", recorder.Code)
	}
	if recorder := signedInCall(t, handler, http.MethodDelete, filesPath+"/hugo.toml", nil); recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("DELETE = %d, want 405", recorder.Code)
	}
}
