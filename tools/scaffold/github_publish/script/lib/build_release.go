package lib

import (
	"github.com/xhd2015/kool/pkgs/release"
)

var DefaultSpecs = release.DefaultSpecs

func BuildRelease(specs []*release.Spec) (*release.BuildReleaseResult, error) {
	// Add custom pre-build steps here (e.g. frontend build, asset generation)
	return release.BuildRelease("__NAME__", nil, specs)
}
