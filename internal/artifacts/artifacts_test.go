package artifacts_test

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/artifacts"
	"github.com/quiqr/chatcms-app/internal/workflows"
)

type memory struct {
	mu   sync.Mutex
	kept []artifacts.Artifact
	fail error
}

func (m *memory) Insert(ctx context.Context, artifact artifacts.Artifact) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.fail != nil {
		return m.fail
	}
	m.kept = append(m.kept, artifact)
	return nil
}

func (m *memory) List(ctx context.Context, projectID uuid.UUID, limit int32) ([]artifacts.Artifact, error) {
	found := []artifacts.Artifact{}
	for i := len(m.kept) - 1; i >= 0; i-- {
		if m.kept[i].ProjectID == projectID {
			found = append(found, m.kept[i])
		}
	}
	return found, nil
}

func (m *memory) Get(ctx context.Context, projectID, id uuid.UUID) (artifacts.Artifact, error) {
	for _, one := range m.kept {
		if one.ProjectID == projectID && one.ID == id {
			return one, nil
		}
	}
	return artifacts.Artifact{}, artifacts.ErrMissing
}

func (m *memory) OfPublish(ctx context.Context, publishID uuid.UUID) ([]artifacts.Artifact, error) {
	found := []artifacts.Artifact{}
	for _, one := range m.kept {
		if one.PublishID != nil && *one.PublishID == publishID {
			found = append(found, one)
		}
	}
	sort.Slice(found, func(a, b int) bool { return found[a].Name < found[b].Name })
	return found, nil
}

func (m *memory) Any(ctx context.Context, projectID uuid.UUID) (bool, error) {
	list, _ := m.List(ctx, projectID, 1)
	return len(list) > 0, nil
}

var built = fstest.MapFS{
	"Korenaar.pdf":         {Data: []byte("%PDF-1.7 the book")},
	"downloads/notes.docx": {Data: []byte("docx")},
}

func TestAPublishKeepsItsFilesAndTheyCanBeOpened(t *testing.T) {
	records := &memory{}
	dir := t.TempDir()
	store, err := artifacts.New(records, dir, 0)
	if err != nil {
		t.Fatal(err)
	}
	run := workflows.Run{ProjectID: uuid.New(), PublishID: uuid.New()}
	kept, err := store.Keep(t.Context(), run, built, []string{"Korenaar.pdf", "downloads/notes.docx"})
	if err != nil || len(kept) != 2 || kept[0].Name != "Korenaar.pdf" || kept[0].Size != 17 {
		t.Fatalf("Keep() = %+v, %v", kept, err)
	}

	artifact, file, err := store.Open(t.Context(), run.ProjectID, kept[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(file)
	file.Close()
	if string(data) != "%PDF-1.7 the book" || artifact.SHA256 == "" || artifact.PublishID == nil || *artifact.PublishID != run.PublishID {
		t.Errorf("Open() = %+v, %q", artifact, data)
	}

	if _, _, err := store.Open(t.Context(), uuid.New(), kept[0].ID); !errors.Is(err, artifacts.ErrMissing) {
		t.Errorf("another project's artifact error = %v", err)
	}
	ofPublish, _ := store.OfPublish(t.Context(), run.PublishID)
	any, _ := store.Any(t.Context(), run.ProjectID)
	none, _ := store.Any(t.Context(), uuid.New())
	if len(ofPublish) != 2 || !any || none {
		t.Errorf("OfPublish() = %d, Any() = %v, %v", len(ofPublish), any, none)
	}
	leftovers, _ := filepath.Glob(filepath.Join(dir, run.ProjectID.String(), ".keeping-*"))
	if len(leftovers) != 0 {
		t.Errorf("half-written files left: %v", leftovers)
	}
}

func TestALaterPublishKeepsANewArtifact(t *testing.T) {
	store, _ := artifacts.New(&memory{}, t.TempDir(), 0)
	project := uuid.New()
	first, _ := store.Keep(t.Context(), workflows.Run{ProjectID: project}, built, []string{"Korenaar.pdf"})
	second, _ := store.Keep(t.Context(), workflows.Run{ProjectID: project}, built, []string{"Korenaar.pdf"})
	list, err := store.List(t.Context(), project)
	if err != nil || len(list) != 2 || list[0].ID != second[0].ID || list[1].ID != first[0].ID {
		t.Errorf("List() = %+v, %v", list, err)
	}
}

func TestAFileTooBigIsNotKept(t *testing.T) {
	records := &memory{}
	store, _ := artifacts.New(records, t.TempDir(), 8)
	_, err := store.Keep(t.Context(), workflows.Run{ProjectID: uuid.New()}, built, []string{"downloads/notes.docx", "Korenaar.pdf"})
	if !errors.Is(err, workflows.ErrTooBig) || !strings.Contains(err.Error(), "Korenaar.pdf") || len(records.kept) != 0 {
		t.Errorf("Keep() error = %v, kept %d", err, len(records.kept))
	}
}

func TestAFileWhoseRecordFailsIsNotLeftBehind(t *testing.T) {
	dir := t.TempDir()
	store, _ := artifacts.New(&memory{fail: errors.New("database down")}, dir, 0)
	project := uuid.New()
	if _, err := store.Keep(t.Context(), workflows.Run{ProjectID: project}, built, []string{"Korenaar.pdf"}); err == nil {
		t.Fatal("Keep() succeeded without a record")
	}
	entries, _ := os.ReadDir(filepath.Join(dir, project.String()))
	if len(entries) != 0 {
		t.Errorf("left %d files", len(entries))
	}
}

func TestAMissingFileIsMissing(t *testing.T) {
	records := &memory{}
	dir := t.TempDir()
	store, _ := artifacts.New(records, dir, 0)
	project := uuid.New()
	kept, _ := store.Keep(t.Context(), workflows.Run{ProjectID: project}, built, []string{"Korenaar.pdf"})
	os.Remove(filepath.Join(dir, project.String(), kept[0].ID.String()))
	if _, _, err := store.Open(t.Context(), project, kept[0].ID); !errors.Is(err, artifacts.ErrMissing) {
		t.Errorf("Open() error = %v", err)
	}
	if _, err := store.Keep(t.Context(), workflows.Run{ProjectID: project}, built, []string{"nothing.pdf"}); err == nil {
		t.Error("a file that is not there was kept")
	}
	if _, err := artifacts.New(nil, dir, 0); !errors.Is(err, artifacts.ErrIncomplete) {
		t.Errorf("New() without records error = %v", err)
	}
}
