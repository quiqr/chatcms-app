package version

import (
	"fmt"
	"runtime"

	chatcms "github.com/quiqr/chatcms-app"
)

var Version string

func init() {
	if Version == "" {
		Version = chatcms.Version()
	}
}

func String() string {
	return fmt.Sprintf("chatcmsd %s (%s/%s, %s)",
		Version, runtime.GOOS, runtime.GOARCH, runtime.Version())
}
