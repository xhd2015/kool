package update

import (
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
