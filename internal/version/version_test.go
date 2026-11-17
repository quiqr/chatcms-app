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

func TestDescribeSaysWhereTheBuildCameFrom(t *testing.T) {
	saved := Source
	t.Cleanup(func() { Source = saved })

	cases := []struct {
		source, want string
		changed      bool
	}{
		{"a3bab43", Version + " · a3bab43", false},
		{"a3bab43-dirty", Version + " · a3bab43, changed", true},
		{"", Version + " · development build", false},
	}
	for _, c := range cases {
		Source = c.source
		if got := Describe(); got != c.want {
			t.Errorf("Describe() with %q = %q, want %q", c.source, got, c.want)
		}
		if got := Changed(); got != c.changed {
			t.Errorf("Changed() with %q = %v, want %v", c.source, got, c.changed)
		}
		if !strings.Contains(String(), c.want) {
			t.Errorf("String() with %q = %q, want it to contain %q", c.source, String(), c.want)
		}
	}
}

func TestTheVersionComesFromTheFile(t *testing.T) {
	if Version != chatcms.Version() {
		t.Errorf("Version = %q, want %q", Version, chatcms.Version())
	}
}
