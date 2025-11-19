package tag

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

var ErrNoTag = fmt.Errorf("no tag found")

func ListTagsAtHEAD(targetDir string) ([]string, error) {
	return ListTags(targetDir, "HEAD")
}
func GetVersionTagAtHEAD(targetDir, tagPrefix string) (string, error) {
	return GetVersionTag(targetDir, "HEAD", tagPrefix)
}

// ListTagsAtHEAD returns all tags pointing at HEAD in the target directory
func ListTags(targetDir string, commit string) ([]string, error) {
	tagOutput, err := cmd.Dir(targetDir).Output("git", "tag", "-l", "--points-at", commit)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags at %s for %s: %w", commit, targetDir, err)
	}

	tags := strings.Split(strings.TrimSpace(string(tagOutput)), "\n")
	if len(tags) == 1 && tags[0] == "" {
		return nil, nil
	}

	// Filter out empty strings
	var result []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			result = append(result, tag)
		}
	}
	return result, nil
}

// GetVersionTagAtHEAD checks if there's a version tag at HEAD in the target directory
// tagPrefix is an optional prefix for the tag (e.g., "path/to/submodule/" for nested modules)
// If tagPrefix is empty, it will match tags like "v0.0.87"
// If tagPrefix is provided, it will match tags like "path/to/submodule/v0.0.87"
func GetVersionTag(targetDir, commit, tagPrefix string) (string, error) {
	// Get all tags pointing at HEAD
	tags, err := ListTags(targetDir, commit)
	if err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("%w: %s in %s, please commit and tag first", ErrNoTag, commit, targetDir)
	}

	// Find a matching version tag
	for _, tag := range tags {
		// If tagPrefix is provided, expect tags like "path/to/submodule/v0.0.87"
		if tagPrefix != "" {
			if strings.HasPrefix(tag, tagPrefix) {
				versionPart := strings.TrimPrefix(tag, tagPrefix)
				if strings.HasPrefix(versionPart, "v") {
					return tag, nil
				}
			}
		} else {
			// For root level, expect tags like "v0.0.87"
			if strings.HasPrefix(tag, "v") {
				return tag, nil
			}
		}
	}

	// No matching version tag found
	if tagPrefix != "" {
		return "", fmt.Errorf("%w: with prefix %s found at %s in %s, please tag with format %sv0.0.X", ErrNoTag, tagPrefix, commit, targetDir, tagPrefix)
	}
	return "", fmt.Errorf("%w: (v0.0.X) found at %s in %s, please tag first", ErrNoTag, commit, targetDir)
}

// GetLatestVersionTag returns the latest version tag in the directory that has versionPrefix as prefix
// The basic name (part after stripping versionPrefix) should not contain "/"
// If versionPrefix is "", then the returned tag should not be nested (no "/" in it)
func GetLatestVersionTag(dir string, versionPrefix string) (string, error) {
	// Get all tags in the repository, sorted by version (latest first)
	tagOutput, err := cmd.Dir(dir).Output("git", "tag", "-l", "--sort=-version:refname", versionPrefix+"*")
	if err != nil {
		return "", fmt.Errorf("failed to get tags for %s: %w", dir, err)
	}

	tags := strings.Split(strings.TrimSpace(string(tagOutput)), "\n")
	if len(tags) == 1 && tags[0] == "" {
		return "", fmt.Errorf("%w: %s", ErrNoTag, dir)
	}

	// Find the latest tag that matches the criteria
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}

		// Check if tag has the required prefix
		if versionPrefix != "" {
			if !strings.HasPrefix(tag, versionPrefix) {
				continue
			}
			// Extract the basic name (part after versionPrefix)
			basicName := strings.TrimPrefix(tag, versionPrefix)
			// Basic name should not contain "/"
			if strings.Contains(basicName, "/") {
				continue
			}
		} else {
			// If versionPrefix is "", tag should not be nested (no "/" in it)
			if strings.Contains(tag, "/") {
				continue
			}
		}

		// This tag matches our criteria
		return tag, nil
	}

	if versionPrefix != "" {
		return "", fmt.Errorf("%w: (%sv0.0.X) in %s", ErrNoTag, versionPrefix, dir)
	}
	return "", fmt.Errorf("%w: (v0.0.X) in %s", ErrNoTag, dir)
}
