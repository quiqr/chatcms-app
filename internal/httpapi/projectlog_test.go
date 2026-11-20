package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"testing"
)

type stubLog struct {
	asked []string
	err   error
}

func (s *stubLog) ProjectLog(_ context.Context, _ Actor, organisation, project, kind, before string) (LogView, error) {
	s.asked = append(s.asked, "log "+organisation+"/"+project+" "+kind+" "+before)
	return LogView{Entries: []LogEntryView{{Entry: "start:7", Kind: "start", State: "started", What: "gvisor"}}, More: true}, s.err
}

func (s *stubLog) LogEntry(_ context.Context, _ Actor, organisation, project, entry string) (LogDetailView, error) {
	s.asked = append(s.asked, "entry "+organisation+"/"+project+" "+entry)
	return LogDetailView{LogEntryView: LogEntryView{Entry: entry}, Hugo: &HugoFactsView{Asked: "0.101.0", Because: "pinned"}}, s.err
}

func withLog(service ProjectLogService, signedIn bool) http.Handler {
	known := map[string]Actor{}
	if signedIn {
		known["a-session-value"] = Actor{ID: "person-1", Email: "anna@example.org"}
	}
	return New(Options{
		Logger:     slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Health:     func(context.Context) Health { return Health{Database: DatabaseReachable} },
		Auth:       &stubAuth{known: known},
		ProjectLog: service,
	})
}

const logPath = "/api/organisations/technative/projects/korenaar/log"

func TestTheLogNeedsAPerson(t *testing.T) {
	handler := withLog(&stubLog{}, false)
	for _, path := range []string{logPath, logPath + "/start:7"} {
		if recorder := call(t, handler, http.MethodGet, path, map[string]any{}); recorder.Code != http.StatusUnauthorized {
			t.Errorf("GET %s = %d, want 401", path, recorder.Code)
		}
	}
}

func TestTheLogPassesItsFiltersOn(t *testing.T) {
	service := &stubLog{}
	handler := withLog(service, true)

	recorder := signedInCall(t, handler, http.MethodGet, logPath+"?kind=start&before=2026-11-20T10:00:00Z", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET = %d: %s", recorder.Code, recorder.Body)
	}
	var view LogView
	if err := json.NewDecoder(recorder.Body).Decode(&view); err != nil {
		t.Fatal(err)
	}
	if !view.More || len(view.Entries) != 1 || view.Entries[0].Entry != "start:7" {
		t.Errorf("view = %+v", view)
	}

	recorder = signedInCall(t, handler, http.MethodGet, logPath+"/start:7", nil)
	var detail LogDetailView
	if err := json.NewDecoder(recorder.Body).Decode(&detail); err != nil || detail.Hugo == nil || detail.Hugo.Asked != "0.101.0" {
		t.Errorf("detail = %+v, %v", detail, err)
	}
	want := []string{"log technative/korenaar start 2026-11-20T10:00:00Z", "entry technative/korenaar start:7"}
	if len(service.asked) != 2 || service.asked[0] != want[0] || service.asked[1] != want[1] {
		t.Errorf("asked = %v", service.asked)
	}
}

func TestTheLogIsRefusedToSomebodyWithoutTechnicalDetail(t *testing.T) {
	handler := withLog(&stubLog{err: ErrNotAllowed}, true)
	if recorder := signedInCall(t, handler, http.MethodGet, logPath, nil); recorder.Code != http.StatusForbidden {
		t.Errorf("GET = %d, want 403", recorder.Code)
	}
	if recorder := signedInCall(t, handler, http.MethodPost, logPath, nil); recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST = %d, want 405", recorder.Code)
	}
}
