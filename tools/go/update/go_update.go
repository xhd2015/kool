package update

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/xhd2015/kool/tools/git/tag"
	"github.com/xhd2015/kool/tools/go/commands"
)

func Update(dir string) error {
	// Check if directory is provided
	if dir == "" {
		return fmt.Errorf("requires dir")
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("no such dir: %s", dir)
	}

	// 	mod=$(cd "$dir" && go mod edit -json|jq -r '.Module.Path')
	// if [[ -z $mod ]];then
	//     echo "not a go module: $dir" >&2
	//     exit 1
	// fi

	opts := &commands.GoModEditOptions{Dir: dir, Stderr: true}
	mod, err := commands.GoModEditJSON(opts)
	if err != nil {
		return fmt.Errorf("failed to get module info: %w", err)
	}
	if mod.Module.Path == "" {
		return fmt.Errorf("not a go module: %s", dir)
	}

	versionPrefix, err := calculateVersionPrefix(dir, mod.Module.Path)
	if err != nil {
		return fmt.Errorf("failed to calculate version prefix for %s: %w", mod.Module.Path, err)
	}
	latestTag, err := tag.GetLatestVersionTag(dir, versionPrefix)
	if err != nil {
		return fmt.Errorf("failed to get latest version tag for %s: %w", mod.Module.Path, err)
	}
	version := tag.StripVersionPrefix(versionPrefix, latestTag)
	if !isValidVersionTag(version) {
		return fmt.Errorf("latest version tag %s resolved to invalid version %s", latestTag, version)
	}

	// do not use go get, not always work
	// Drop the replacement first, then update the version
	if err := commands.GoModDropReplace(mod.Module.Path, nil); err != nil {
		return err
	}
	if err := commands.GoModEditRequire(mod.Module.Path, version, nil); err != nil {
		return err
	}

	msgCmd := exec.Command("git", "log", "-1", "--format=%s", latestTag)
	msgCmd.Dir = dir
	msgCmd.Stderr = os.Stderr
	msgOutput, err := msgCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get commit message: %v", err)
	}
	msg := strings.TrimSpace(string(msgOutput))
	fmt.Printf("commit message: %s\n", msg)
	return nil
}

// stripSubDirFromTag strips the sub-directory prefix from a tag if the module path ends with that prefix
// For example:
//
//	tag: "sub/module-a/v1.20.1", modulePath: "github.com/example/repo/sub/module-a" -> "v1.20.1"
//	tag: "v1.20.1", modulePath: "github.com/example/repo" -> "v1.20.1" (unchanged)
func stripSubDirFromTag(tag, modulePath string) string {
	if tag == "" {
		return tag
	}

	// Split the tag by "/" to find potential sub-directory prefix
	tagParts := strings.Split(tag, "/")
	if len(tagParts) <= 1 {
		// No sub-directory in tag, return as-is
		return tag
	}

	// The last part should be the actual version (e.g., "v1.20.1")
	version := tagParts[len(tagParts)-1]

	// The prefix parts are the sub-directory (e.g., ["sub", "module-a"])
	subDirParts := tagParts[:len(tagParts)-1]
	subDirPath := strings.Join(subDirParts, "/")

	// Check if the module path ends with this sub-directory path
	if strings.HasSuffix(modulePath, "/"+subDirPath) || strings.HasSuffix(modulePath, subDirPath) {
		// Strip the sub-directory prefix and return just the version
		return version
	}

	// If module path doesn't match, return the original tag
	return tag
}
