package tag

import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/gitops/git"
)

func GetVersionPrefix(dir string) (string, error) {
	_, subPathList, err := GetSubPath(dir)
	if err != nil {
		return "", err
	}
	if len(subPathList) == 0 {
		return "v", err
	}
	return strings.Join(subPathList, "/") + "/v", nil
}

func AddVersionPrefix(prefix string) string {
	if prefix == "" {
		return "v"
	}
	return strings.TrimSuffix(prefix, "/") + "/v"
}

// StripVersionPrefix strips the version prefix from a tag to get the clean version
// Examples:
//   - StripVersionPrefix("", "v1.2.3") -> "v1.2.3"
//   - StripVersionPrefix("submodule/", "submodule/v1.2.3") -> "v1.2.3"
//   - StripVersionPrefix("path/to/module/", "path/to/module/v2.0.0") -> "v2.0.0"
func StripVersionPrefix(prefix string, tag string) string {
	version := tag
	if strings.HasPrefix(tag, prefix) {
		version = tag[len(prefix):]
	}
	return "v" + strings.TrimPrefix(version, "v")
}

func GetSubPath(dir string) (string, []string, error) {
	topLevel, err := git.ShowToplevel(dir)
	if err != nil {
		return "", nil, err
	}

	absTopLevel, err := filepath.Abs(topLevel)
	if err != nil {
		return "", nil, err
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", nil, err
	}

	subPath, err := filepath.Rel(absTopLevel, absDir)
	if err != nil {
		return "", nil, err
	}
	var subPathList []string
	if subPath != "" && subPath != "." {
		subPathList = filepath.SplitList(subPath)
	}

	return topLevel, subPathList, nil
}
