package version

import (
	// embed is used to embed the VERSION file into the binary
	_ "embed"
	"fmt"
	"strings"

	version "github.com/hashicorp/go-version"
)

// rawVersion is read from the flat VERSION file. This must be a valid semantic version string.
//
//go:embed VERSION
var rawVersion string

// dev determines whether the -dev prerelease marker will
// be included in version info. It is expected to be set to "no" using
// linker flags when building binaries for release.
var dev string = "yes"

// gitHeadChecksum is the git commit hash that was compiled.
// This will be populated with linker flags when building dev binaries.
var gitHeadChecksum string

// Version stores the main version number for this binary, populated from the raw version.
var Version string

// Prerelease stores an optional version string, populated using a combination of the raw version
// and the dev flag.
var Prerelease string

// SemVer is an instance of version.Version representing the main version
// without any prerelease information.
var SemVer *version.Version

func init() {
	semVerFull := version.Must(version.NewVersion(strings.TrimSpace(rawVersion)))
	SemVer = semVerFull.Core()
	Version = SemVer.String()

	if dev == "no" {
		Prerelease = semVerFull.Prerelease()
	} else {
		if gitHeadChecksum != "" {
			Prerelease = "dev-" + gitHeadChecksum
		} else {
			Prerelease = "dev"
		}
	}
}

// String returns the complete version string, including prerelease
func String() string {
	if Prerelease != "" {
		return fmt.Sprintf("%s-%s", Version, Prerelease)
	}
	return Version
}
