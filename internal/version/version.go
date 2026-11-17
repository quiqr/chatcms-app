package version

import (
	"fmt"
	"runtime"
	"strings"

	chatcms "github.com/quiqr/chatcms-app"
)

var Version string

var Source string

const changedSuffix = "-dirty"

func init() {
	if Version == "" {
		Version = chatcms.Version()
	}
}

func Changed() bool {
	return strings.HasSuffix(Source, changedSuffix)
}

func Short() string {
	return strings.TrimSuffix(Source, changedSuffix)
}

func Describe() string {
	switch {
	case Short() == "":
		return Version + " · development build"
	case Changed():
		return Version + " · " + Short() + ", changed"
	default:
		return Version + " · " + Short()
	}
}

func String() string {
	return fmt.Sprintf("chatcmsd %s (%s/%s, %s)",
		Describe(), runtime.GOOS, runtime.GOARCH, runtime.Version())
}
