package update

import (
	"strings"

	"github.com/Masterminds/semver/v3"
)

// isValidVersionTag checks if a version represents a valid semantic version
func isValidVersionTag(version string) bool {
	if version == "" {
		return false
	}

	// Parse with semver library
	_, err := semver.NewVersion(version)
	return err == nil
}

// isNewerVersion compares two versions and returns true if version 'a' is newer than version 'b'
func isNewerVersion(a, b string) bool {
	if a == "" || b == "" {
		// If either version is empty, return false
		return false
	}

	semverA, errA := semver.NewVersion(a)
	semverB, errB := semver.NewVersion(b)

	if errA != nil || errB != nil {
		// If either version is invalid, return false
		return false
	}

	return semverA.GreaterThan(semverB)
}

// stripVersionTagPrefix strips the version prefix from a tag to get the clean version
// Examples:
//   - stripVersionTagPrefix("", "v1.2.3") -> "v1.2.3"
//   - stripVersionTagPrefix("submodule/", "submodule/v1.2.3") -> "v1.2.3"
//   - stripVersionTagPrefix("path/to/module/", "path/to/module/v2.0.0") -> "v2.0.0"
func stripVersionTagPrefix(tagPrefix string, tag string) string {
	if tagPrefix == "" {
		return tag
	}
	if strings.HasPrefix(tag, tagPrefix) {
		return strings.TrimPrefix(tag, tagPrefix)
	}
	// If tag doesn't have the expected prefix, return as-is
	return tag
}
