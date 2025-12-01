package config

import "fmt"

// The following vars are automatically injected via -ldflags.
// See Makefile target "make go-build" and make var $(LDFLAGS).
// No need to change them here.
// https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications
var (
	// ModuleName is the name of the Go module, typically "github.com/pmaojo/goploy".
	// It is injected via ldflags during the build process.
	ModuleName = "build.local/misses/ldflags"
	// Commit is the 40-character Git commit hash of the current build.
	// It is injected via ldflags during the build process.
	Commit = "< 40 chars git commit hash via ldflags >"
	// BuildDate is the ISO 8601 timestamp of when the binary was built.
	// It is injected via ldflags during the build process.
	BuildDate = "1970-01-01T00:00:00+00:00"
)

// GetFormattedBuildArgs returns a formatted string containing the module name, commit hash, and build date.
//
// Returns:
//   A string in the format "<ModuleName> @ <Commit> (<BuildDate>)"
func GetFormattedBuildArgs() string {
	return fmt.Sprintf("%v @ %v (%v)", ModuleName, Commit, BuildDate)
}
