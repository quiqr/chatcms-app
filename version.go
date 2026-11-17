package chatcms

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var versionFile string

//go:embed LICENSE
var licenseFile string

func License() string {
	return licenseFile
}

func Version() string {
	return strings.TrimSpace(versionFile)
}
