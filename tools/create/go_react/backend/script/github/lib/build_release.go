// usage: imported by go run ./script/github/release (shared release helpers)
//
// Proposed behavior (sketch):
//   1. Expose DefaultSpecs for multi-platform release builds.
//   2. BuildRelease runs optional pre-build steps then release.BuildRelease.
//   3. Callers pass specs; name/module placeholders are substituted at scaffold time.
package lib

import (
	"github.com/xhd2015/kool/pkgs/release"
)

var DefaultSpecs = release.DefaultSpecs

func BuildRelease(specs []*release.Spec) (*release.BuildReleaseResult, error) {
	// Add custom pre-build steps here (e.g. frontend build, asset generation)
	return release.BuildRelease("__PROJECT_NAME__", nil, specs)
}
