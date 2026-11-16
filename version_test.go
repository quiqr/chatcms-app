package chatcms

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestVersionIsTheFileTrimmed(t *testing.T) {
	raw, err := os.ReadFile("VERSION")
	if err != nil {
		t.Fatal(err)
	}

	if got, want := Version(), strings.TrimSpace(string(raw)); got != want {
		t.Errorf("Version() = %q, want %q", got, want)
	}
}

func TestVersionIsSemver(t *testing.T) {
	if !regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`).MatchString(Version()) {
		t.Errorf("Version() = %q, want a bare semver", Version())
	}
}
