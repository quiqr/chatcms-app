//go:build integration

package artifacts_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/artifacts"
	"github.com/quiqr/chatcms-app/internal/store"
	"github.com/quiqr/chatcms-app/internal/testsupport"
	"github.com/quiqr/chatcms-app/internal/workflows"
)

func TestPostgresKeepsArtifactsPerProjectAndPublish(t *testing.T) {
	db, err := store.Open(t.Context(), testsupport.Postgres(t))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}
	organisation, err := db.Queries.CreateOrganisation(t.Context(), store.CreateOrganisationParams{Slug: "korenaar", Name: "Korenaar"})
	if err != nil {
		t.Fatal(err)
	}
	project, err := db.Queries.CreateProject(t.Context(), store.CreateProjectParams{OrganisationID: organisation.ID, Slug: "book", Name: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	publish, err := db.Queries.StartPublish(t.Context(), store.StartPublishParams{ProjectID: project.ID, Summary: []byte("[]")})
	if err != nil {
		t.Fatal(err)
	}
	projectID, publishID := uuid.UUID(project.ID.Bytes), uuid.UUID(publish.ID.Bytes)

	kept, err := artifacts.New(artifacts.NewPostgres(db.Queries), t.TempDir(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if any, _ := kept.Any(t.Context(), projectID); any {
		t.Error("a new project has artifacts")
	}
	made, err := kept.Keep(t.Context(), workflows.Run{ProjectID: projectID, PublishID: publishID}, built, []string{"Korenaar.pdf", "downloads/notes.docx"})
	if err != nil {
		t.Fatal(err)
	}
	unattached, err := kept.Keep(t.Context(), workflows.Run{ProjectID: projectID}, built, []string{"Korenaar.pdf"})
	if err != nil {
		t.Fatal(err)
	}

	list, err := kept.List(t.Context(), projectID)
	if err != nil || len(list) != 3 {
		t.Fatalf("List() = %+v, %v", list, err)
	}
	ofPublish, err := kept.OfPublish(t.Context(), publishID)
	if err != nil || len(ofPublish) != 2 || ofPublish[0].Name != "Korenaar.pdf" || ofPublish[1].Name != "downloads/notes.docx" {
		t.Errorf("OfPublish() = %+v, %v", ofPublish, err)
	}
	got, file, err := kept.Open(t.Context(), projectID, made[0].ID)
	if err != nil || got.Size != 17 || got.PublishID == nil || *got.PublishID != publishID {
		t.Errorf("Open() = %+v, %v", got, err)
	}
	file.Close()
	if loose, file, err := kept.Open(t.Context(), projectID, unattached[0].ID); err != nil || loose.PublishID != nil {
		t.Errorf("an artifact of no publish = %+v, %v", loose, err)
	} else {
		file.Close()
	}
	if _, _, err := kept.Open(t.Context(), uuid.New(), made[0].ID); err == nil {
		t.Error("another project opened the artifact")
	}
	if any, _ := kept.Any(t.Context(), projectID); !any {
		t.Error("Any() = false after keeping")
	}
}
