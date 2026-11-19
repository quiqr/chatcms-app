package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestShortcodesAreDeclaredInTheBaseAndInIncludes(t *testing.T) {
	m, err := Load(os.DirFS(filepath.Join(fixturesDir, "shortcodes")))
	if err != nil {
		t.Fatal(err)
	}
	if m.ShortcodeDialect != "hugo" {
		t.Fatalf("dialect = %q", m.ShortcodeDialect)
	}
	var keys []string
	for _, shortcode := range m.Shortcodes {
		keys = append(keys, shortcode.Key)
	}
	if got := strings.Join(keys, ","); got != "figure,icon,notice" {
		t.Fatalf("shortcodes = %s", got)
	}
	notice := m.Shortcodes[2]
	if notice.Inner != "markdown" || notice.Fields[0].Position == nil || *notice.Fields[0].Position != 0 || notice.Fields[0].Extra != nil {
		t.Fatalf("notice = %+v", notice)
	}
	if m.Shortcodes[0].Fields[0].Type != "image-select" || m.Shortcodes[0].Fields[0].Position != nil {
		t.Fatalf("figure = %+v", m.Shortcodes[0])
	}
}

func TestAModelWithoutShortcodesReadsAsBefore(t *testing.T) {
	m, err := Load(fstest.MapFS{"quiqr/model/base.yaml": {Data: []byte("singles:\n  - key: home\n    file: content/_index.md\n")}})
	if err != nil {
		t.Fatal(err)
	}
	if m.Shortcodes != nil || m.ShortcodeDialect != "" {
		t.Fatalf("shortcodes = %+v, dialect = %q", m.Shortcodes, m.ShortcodeDialect)
	}
}

func TestAMistakeInAShortcodeSaysWhere(t *testing.T) {
	for name, base := range map[string]string{
		"a list field":       "shortcodes:\n  - key: x\n    fields:\n      - key: a\n        type: leaf-array\n",
		"a shared position":  "shortcodes:\n  - key: x\n    fields:\n      - key: a\n        type: string\n        position: 0\n      - key: b\n        type: string\n        position: 0\n",
		"a gap":              "shortcodes:\n  - key: x\n    fields:\n      - key: a\n        type: string\n        position: 1\n",
		"inner html":         "shortcodes:\n  - key: x\n    inner: html\n",
		"an unknown dialect": "shortcodeDialect: textile\n",
		"a shared key":       "shortcodes:\n  - key: x\n  - key: x\n",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Load(fstest.MapFS{"quiqr/model/base.yaml": {Data: []byte(base)}})
			if err == nil {
				t.Fatal("the mistake was taken without a word")
			}
			if !strings.Contains(err.Error(), "base.yaml") {
				t.Fatalf("the error does not say where: %v", err)
			}
		})
	}
}
