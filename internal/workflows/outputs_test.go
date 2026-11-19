package workflows_test

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/runner"
	"github.com/quiqr/chatcms-app/internal/workflows"
)

const book = `on: publish
steps:
  - action: build
  - action: publish
outputs:
  - path: .
    to: publish
  - path: "*.pdf"
    to: artifact
`

func TestOutputsAreRead(t *testing.T) {
	workflow, err := workflows.Parse("quiqr/workflows/publish.yaml", []byte(book))
	if err != nil {
		t.Fatal(err)
	}
	if len(workflow.Outputs) != 2 || workflow.Outputs[0] != (workflows.Output{Path: ".", To: workflows.ToPublish}) || workflow.Outputs[1].To != workflows.ToArtifact {
		t.Errorf("Outputs = %+v", workflow.Outputs)
	}
}

func TestOutputsThatCannotWorkAreRefused(t *testing.T) {
	for name, tc := range map[string]struct{ file, says string }{
		"no build":         {"on: publish\nsteps:\n  - action: publish\noutputs:\n  - path: .\n    to: publish\n", "does not build"},
		"no publish":       {"on: publish\nsteps:\n  - action: build\noutputs:\n  - path: .\n    to: both\n", "does not publish"},
		"no destination":   {"on: publish\nsteps:\n  - action: build\noutputs:\n  - path: x.pdf\n    to: cloud\n", "artifact, both, publish"},
		"no path":          {"on: publish\nsteps:\n  - action: build\noutputs:\n  - to: artifact\n", "which part"},
		"outside":          {"on: publish\nsteps:\n  - action: build\noutputs:\n  - path: ../secret\n    to: artifact\n", "inside the build"},
		"absolute":         {"on: publish\nsteps:\n  - action: build\noutputs:\n  - path: /etc/passwd\n    to: artifact\n", "inside the build"},
		"bad pattern":      {"on: publish\nsteps:\n  - action: build\noutputs:\n  - path: \"[x\"\n    to: artifact\n", "not a pattern"},
		"misspelled field": {"on: publish\nsteps:\n  - action: build\noutputs:\n  - path: x\n    too: artifact\n", "unknown"},
	} {
		_, err := workflows.Parse("publish.yaml", []byte(tc.file))
		var wrong *workflows.Error
		if !errors.As(err, &wrong) || !strings.Contains(err.Error(), tc.says) {
			t.Errorf("%s: error = %v, want it to say %q", name, err, tc.says)
		}
	}
}

func TestAnOutputHoldsFilesDirectoriesAndPatterns(t *testing.T) {
	for _, tc := range []struct {
		output string
		file   string
		holds  bool
	}{
		{".", "a/b.html", true},
		{"book.pdf", "book.pdf", true},
		{"downloads", "downloads/book.pdf", true},
		{"downloads", "downloadsx/book.pdf", false},
		{"*.pdf", "book.pdf", true},
		{"*.pdf", "chapters/book.pdf", false},
		{"chapters/*.html", "chapters/bread.html", true},
	} {
		if got := (workflows.Output{Path: tc.output, To: workflows.ToArtifact}).Holds(tc.file); got != tc.holds {
			t.Errorf("%q holds %q = %v", tc.output, tc.file, got)
		}
	}
}

type keeper struct {
	run   workflows.Run
	names []string
	read  map[string]string
	err   error
}

func (k *keeper) Keep(ctx context.Context, run workflows.Run, files fs.FS, names []string) ([]workflows.Kept, error) {
	k.run, k.names, k.read = run, names, map[string]string{}
	for _, name := range names {
		data, _ := fs.ReadFile(files, name)
		k.read[name] = string(data)
	}
	if k.err != nil {
		return nil, k.err
	}
	kept := []workflows.Kept{}
	for _, name := range names {
		kept = append(kept, workflows.Kept{ID: uuid.New(), Name: name, Size: int64(len(k.read[name]))})
	}
	return kept, nil
}

func builtBook() *recordedBuilder {
	return &recordedBuilder{result: &runner.Result{Root: "_book", Outputs: fstest.MapFS{
		"_book/index.html":          {Data: []byte("<h1>Korenaar</h1>")},
		"_book/chapters/bread.html": {Data: []byte("bread")},
		"_book/Korenaar.pdf":        {Data: []byte("%PDF-")},
	}}}
}

func TestABookPublishesItsSiteAndKeepsItsPDF(t *testing.T) {
	workflow, _ := workflows.Parse("publish.yaml", []byte(strings.Replace(book, "to: artifact", "to: artifact\n  - path: index.html\n    to: both", 1)))
	workflow.Outputs[0].Path = "chapters"
	publisher := &recordedPublisher{}
	kept := &keeper{}
	made := engine(t, &recordedRunner{}, builtBook(), publisher)
	made.Keeping(kept)

	run := aRun()
	run.PublishID = uuid.New()
	outcome, err := made.Run(t.Context(), workflow, run)
	if err != nil || !outcome.Succeeded() {
		t.Fatalf("Run() = %v, %v", outcome.Sentence(), err)
	}
	published := []string{}
	fs.WalkDir(publisher.given, ".", func(name string, entry fs.DirEntry, err error) error {
		if err == nil && !entry.IsDir() {
			published = append(published, name)
		}
		return nil
	})
	if strings.Join(published, ",") != "chapters/bread.html,index.html" {
		t.Errorf("published %v", published)
	}
	if strings.Join(kept.names, ",") != "Korenaar.pdf,index.html" || kept.read["Korenaar.pdf"] != "%PDF-" || kept.run.PublishID != run.PublishID {
		t.Errorf("kept %v %v", kept.names, kept.read)
	}
	if len(outcome.Kept) != 2 || outcome.Kept[0].Name != "Korenaar.pdf" {
		t.Errorf("outcome.Kept = %+v", outcome.Kept)
	}
}

func TestWithoutOutputsTheWholeBuildIsPublishedAndNothingKept(t *testing.T) {
	publisher := &recordedPublisher{}
	kept := &keeper{}
	made := engine(t, &recordedRunner{}, builtBook(), publisher)
	made.Keeping(kept)
	workflow := workflows.Workflow{Trigger: workflows.OnPublish, Steps: []workflows.Step{{Action: workflows.Build}, {Action: workflows.Publish}}}
	if _, err := made.Run(t.Context(), workflow, aRun()); err != nil {
		t.Fatal(err)
	}
	if _, err := fs.Stat(publisher.given, "Korenaar.pdf"); err != nil || kept.names != nil {
		t.Errorf("published %v, kept %v", err, kept.names)
	}
}

func TestNothingIsKeptWhenAStepFails(t *testing.T) {
	workflow, _ := workflows.Parse("publish.yaml", []byte(book))
	kept := &keeper{}
	made := engine(t, &recordedRunner{}, builtBook(), &recordedPublisher{err: errors.New("unreachable")})
	made.Keeping(kept)
	outcome, err := made.Run(t.Context(), workflow, aRun())
	if err != nil || outcome.Succeeded() || kept.names != nil {
		t.Errorf("Run() = %v, %v, kept %v", outcome.Sentence(), err, kept.names)
	}
}

func TestAFileTooBigToKeepIsSaid(t *testing.T) {
	workflow, _ := workflows.Parse("publish.yaml", []byte(book))
	made := engine(t, &recordedRunner{}, builtBook(), &recordedPublisher{})
	made.Keeping(&keeper{err: workflows.ErrTooBig})
	outcome, err := made.Run(t.Context(), workflow, aRun())
	if err != nil || outcome.Succeeded() || !strings.Contains(outcome.Sentence(), "too big") {
		t.Errorf("Run() = %q, %v", outcome.Sentence(), err)
	}

	made.Keeping(&keeper{err: errors.New("disk on fire")})
	if _, err := made.Run(t.Context(), workflow, aRun()); err == nil {
		t.Error("a keeper's failure was swallowed")
	}
}

func TestArtifactsWithoutAKeeperAreAnError(t *testing.T) {
	workflow, _ := workflows.Parse("publish.yaml", []byte(book))
	made := engine(t, &recordedRunner{}, builtBook(), &recordedPublisher{})
	if _, err := made.Run(t.Context(), workflow, aRun()); err == nil {
		t.Error("artifacts vanished without a keeper")
	}
}

func TestTheQuartoBookPublishesItsSiteAndKeepsItsPDF(t *testing.T) {
	set, err := workflows.Load(os.DirFS("../../test/fixtures/quarto-book"))
	if err != nil {
		t.Fatal(err)
	}
	publish := set[workflows.OnPublish]
	if !publish.Selects("Korenaar-Recipes.pdf", workflows.Destination.Keeps) || !publish.Selects("chapters/bread.html", workflows.Destination.Publishes) || publish.Selects("index.html", workflows.Destination.Keeps) {
		t.Errorf("outputs = %+v", publish.Outputs)
	}
}
