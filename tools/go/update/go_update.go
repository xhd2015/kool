package update

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/xhd2015/kool/tools/git/git_tag_next"
	"github.com/xhd2015/kool/tools/git/tag"
	"github.com/xhd2015/kool/tools/go/commands"
	"github.com/xhd2015/kool/tools/go/resolve"
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

	// try best effort to get the tag
	tags, err := tag.ListTagsAtHEAD(dir)
	if err != nil {
		return fmt.Errorf("failed to get tag: %v", err)
	}
	var tagStr string
	if len(tags) > 0 {
		tagStr = tags[0]
	}
	var usePlainReplace bool
	var plainReplaceWith string
	gitRef := tagStr
	if tagStr == "" {
		commitHash, _ := git_tag_next.ShowHeadCommitHash(dir)
		if commitHash != "" {
			gitRef = commitHash
			resolvedTag, _ := resolve.GoResolveVersion(dir, mod.Module.Path, commitHash)
			if resolvedTag != "" {
				tagStr = resolvedTag
			} else {
				usePlainReplace = true
				plainReplaceWith = commitHash
			}
		}
	}
	if tagStr == "" && !usePlainReplace {
		return fmt.Errorf("no tag at HEAD: %s", dir)
	}

	// do not use go get, not always work
	t := time.Now()
	var PLACEHOLDER = fmt.Sprintf("v1.0.%s%d-FAKE", t.Format("20060102150405"), t.Nanosecond())
	editRef := stripSubDirFromTag(tagStr, mod.Module.Path)
	if usePlainReplace {
		editRef = PLACEHOLDER
	}

	// Drop the replacement first, then update the version
	if err := commands.GoModDropReplace(mod.Module.Path, nil); err != nil {
		return err
	}
	if err := commands.GoModEditRequire(mod.Module.Path, editRef, nil); err != nil {
		return err
	}

	if usePlainReplace {
		modFile, err := os.ReadFile("go.mod")
		if err != nil {
			return fmt.Errorf("failed to read go mod: %v", err)
		}
		fmt.Fprintf(os.Stderr, "replace %s %s => %s\n", mod.Module.Path, PLACEHOLDER, plainReplaceWith)
		newModFile := strings.ReplaceAll(string(modFile), PLACEHOLDER, plainReplaceWith)
		err = os.WriteFile("go.mod", []byte(newModFile), 0644)
		if err != nil {
			return fmt.Errorf("failed to write go mod: %v", err)
		}
	}

	msgCmd := exec.Command("git", "log", "-1", "--format=%s", gitRef)
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
