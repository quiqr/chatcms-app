package app

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestARunTimeIsSaidForEachKindOfSite(t *testing.T) {
	quarto := fstest.MapFS{"quiqr/runtime.yaml": {Data: []byte("packages: {quarto: \"1.10.18\"}\npreview: {command: [quarto, preview]}\n")}}
	hugo := fstest.MapFS{"hugo.toml": {Data: []byte("title = 'x'\n")}}
	broken := fstest.MapFS{"quiqr/runtime.yaml": {Data: []byte("preview: {type: video}\n")}}
	for name, tc := range map[string]struct {
		site fstest.MapFS
		want string
	}{"quarto": {quarto, "quarto"}, "hugo": {hugo, "hugo"}, "broken": {broken, ""}, "empty": {fstest.MapFS{}, ""}} {
		if got := runtimeOfSite(tc.site); got != tc.want {
			t.Errorf("%s: runtimeOfSite() = %q, want %q", name, got, tc.want)
		}
	}
	if got := runtimeOfSite(os.DirFS(filepath.Join(t.TempDir(), "no-site-yet"))); got != "" {
		t.Errorf("a project without a site = %q", got)
	}
}
