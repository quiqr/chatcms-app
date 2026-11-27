package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func (s *stubFiles) OpenProjectFile(_ context.Context, _ Actor, organisation, project, path string) (FileContent, error) {
	s.asked = append(s.asked, "open "+path)
	if s.err != nil {
		return FileContent{}, s.err
	}
	preview, kind := "image", "image/png"
	if strings.HasSuffix(path, ".pdf") {
		preview, kind = "pdf", "application/pdf"
	}
	return FileContent{
		Name:     "loaf.png",
		Preview:  preview,
		Type:     kind,
		Modified: time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC),
		Content:  strings.NewReader("0123456789"),
		Close:    func() error { s.asked = append(s.asked, "closed"); return nil },
	}, nil
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

const contentPath = "/api/organisations/technative/projects/korenaar/file-content"

func TestAFileIsServedAsItself(t *testing.T) {
	service := &stubFiles{}
	handler := withFiles(service, true)

	recorder := signedInCall(t, handler, http.MethodGet, contentPath+"/static/images/loaf.png", nil)
	if recorder.Code != http.StatusOK || recorder.Body.String() != "0123456789" {
		t.Fatalf("GET = %d %q", recorder.Code, recorder.Body)
	}
	for header, want := range map[string]string{
		"Content-Type":            "image/png",
		"X-Content-Type-Options":  "nosniff",
		"Content-Security-Policy": "sandbox",
		"Content-Disposition":     "inline; filename=loaf.png",
	} {
		if got := recorder.Header().Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
	if len(service.asked) != 2 || service.asked[0] != "open static/images/loaf.png" || service.asked[1] != "closed" {
		t.Errorf("asked = %v", service.asked)
	}
}

func TestAFileIsServedInRanges(t *testing.T) {
	handler := withFiles(&stubFiles{}, true)
	request := httptest.NewRequest(http.MethodGet, contentPath+"/static/film.mp4", nil)
	request.AddCookie(&http.Cookie{Name: SessionCookie, Value: "a-session-value"})
	request.Header.Set("Range", "bytes=2-4")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusPartialContent || recorder.Body.String() != "234" {
		t.Errorf("range = %d %q", recorder.Code, recorder.Body)
	}
}

func TestAPDFIsServedWithoutASandbox(t *testing.T) {
	recorder := signedInCall(t, withFiles(&stubFiles{}, true), http.MethodGet, contentPath+"/docs/offer.pdf", nil)
	if recorder.Code != http.StatusOK || recorder.Header().Get("Content-Type") != "application/pdf" {
		t.Fatalf("GET = %d %v", recorder.Code, recorder.Header())
	}
	if recorder.Header().Get("Content-Security-Policy") != "" || recorder.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("headers = %v", recorder.Header())
	}
}

func TestAFileThatIsNotShownIsRefused(t *testing.T) {
	service := &stubFiles{err: &RefusedError{Key: "refused.file-not-shown", Reason: "no"}}
	recorder := signedInCall(t, withFiles(service, true), http.MethodGet, contentPath+"/layouts/index.html", nil)
	if recorder.Code < 400 || recorder.Code >= 500 {
		t.Errorf("GET = %d", recorder.Code)
	}
	if recorder := call(t, withFiles(&stubFiles{}, false), http.MethodGet, contentPath+"/a.png", nil); recorder.Code != http.StatusUnauthorized {
		t.Errorf("signed out = %d", recorder.Code)
	}
}
