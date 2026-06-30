package update

import (
	"strings"

	gotoolupdate "github.com/xhd2015/dot-pkgs/go-pkgs/gotool/update"
)

func Update(dir string) error {
	return gotoolupdate.Update(dir)
}

// stripSubDirFromTag strips the sub-directory prefix from a tag if the module path ends with that prefix.
func stripSubDirFromTag(tag, modulePath string) string {
	if tag == "" {
		return tag
	}

	tagParts := strings.Split(tag, "/")
	if len(tagParts) <= 1 {
		return tag
	}

	version := tagParts[len(tagParts)-1]
	subDirParts := tagParts[:len(tagParts)-1]
	subDirPath := strings.Join(subDirParts, "/")

	if strings.HasSuffix(modulePath, "/"+subDirPath) || strings.HasSuffix(modulePath, subDirPath) {
		return version
	}

	return tag
}