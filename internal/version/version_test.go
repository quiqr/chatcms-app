package version

import (
	"runtime"
	"strings"
	"testing"

	chatcms "github.com/quiqr/chatcms-app"
)

func TestStringCarriesVersionAndPlatform(t *testing.T) {
	got := String()

	for _, want := range []string{"chatcmsd", Version, runtime.GOOS, runtime.GOARCH} {
		if !strings.Contains(got, want) {
			t.Errorf("String() = %q, want it to contain %q", got, want)
		}
	}
}

func TestVersionDefaultsToTheVersionFile(t *testing.T) {
	if Version != chatcms.Version() {
		t.Errorf("Version = %q, want the VERSION file's %q", Version, chatcms.Version())
	}
	if !strings.Contains(String(), chatcms.Version()) {
		t.Errorf("String() = %q, want it to carry %q", String(), chatcms.Version())
	}
}
