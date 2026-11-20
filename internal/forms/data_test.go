package forms

import (
	"testing"

	"github.com/quiqr/chatcms-app/internal/model"
)

func TestADataFileIsEditedWholeInItsOwnFormat(t *testing.T) {
	fields := []model.Field{{Key: "note", Type: "string"}, {Key: "open", Type: "boolean"}}
	for name, test := range map[string]struct{ before, after string }{
		"data/hours.yaml": {"note: Closed on Sundays\nopen: true\n", "note: Open on Sundays\nopen: true\n"},
		"config.toml":     {"note = \"Closed on Sundays\"\nopen = true", "note = \"Open on Sundays\"\nopen = true"},
		"data/hours.json": {"{\n  \"note\": \"Closed on Sundays\",\n  \"open\": true\n}", "{\n  \"note\": \"Open on Sundays\",\n  \"open\": true\n}"},
	} {
		form, err := Load(fields, name, []byte(test.before))
		if err != nil {
			t.Fatalf("%s: Load() error = %v", name, err)
		}
		if value, ok := form.Get("note"); !ok || value != "Closed on Sundays" {
			t.Errorf("%s: note = %#v, %v", name, value, ok)
		}
		if err := form.Set("note", "Open on Sundays"); err != nil {
			t.Fatal(err)
		}
		written, err := form.Bytes()
		if err != nil {
			t.Fatalf("%s: Bytes() error = %v", name, err)
		}
		if string(written) != test.after {
			t.Errorf("%s wrote:\n%s\nwant:\n%s", name, written, test.after)
		}
	}
}

func TestAPageKeepsItsFrontMatter(t *testing.T) {
	form, err := Load([]model.Field{{Key: "title", Type: "string"}}, "content/about.md", []byte("---\ntitle: About\n---\n\nText.\n"))
	if err != nil {
		t.Fatal(err)
	}
	written, _ := form.Bytes()
	if string(written) != "---\ntitle: About\n---\n\nText.\n" {
		t.Errorf("wrote %q", written)
	}
}
