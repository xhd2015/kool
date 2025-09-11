package update

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/xhd2015/kool/tools/git/git_tag_next"
	"github.com/xhd2015/kool/tools/go/resolve"
)

type GoMod struct {
	Module struct {
		Path string `json:"Path"`
	} `json:"Module"`
}

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

	modCmd := exec.Command("go", "mod", "edit", "-json")
	modCmd.Dir = dir
	modCmd.Stderr = os.Stderr
	modOutput, err := modCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get module path: %v", err)
	}
	var mod GoMod
	if err := json.Unmarshal(modOutput, &mod); err != nil {
		return fmt.Errorf("failed to unmarshal module path: %v", err)
	}
	if mod.Module.Path == "" {
		return fmt.Errorf("not a go module: %s", dir)
	}

	// try best effort to get the tag
	tagCmd := exec.Command("git", "tag", "-l", "--points-at", "HEAD")
	tagCmd.Dir = dir
	tagCmd.Stderr = os.Stderr
	tagOutput, err := tagCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get tag: %v", err)
	}
	tag := strings.TrimSpace(string(tagOutput))
	var usePlainReplace bool
	var plainReplaceWith string
	gitRef := tag
	if tag == "" {
		commitHash, _ := git_tag_next.ShowHeadCommitHash(dir)
		if commitHash != "" {
			gitRef = commitHash
			resolvedTag, _ := resolve.GoResolveVersion(dir, mod.Module.Path, commitHash)
			if resolvedTag != "" {
				tag = resolvedTag
			} else {
				usePlainReplace = true
				plainReplaceWith = commitHash
			}
		}
	}
	if tag == "" && !usePlainReplace {
		return fmt.Errorf("no tag at HEAD: %s", dir)
	}

	dropArgs := []string{"mod", "edit", "-dropreplace", mod.Module.Path}
	fmt.Fprintf(os.Stderr, "go %s\n", strings.Join(dropArgs, " "))
	editCmd := exec.Command("go", dropArgs...)
	editCmd.Stderr = os.Stderr
	editCmd.Stdout = os.Stdout
	err = editCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to edit go mod: %v", err)
	}

	// do not use go get, not always work
	t := time.Now()
	var PLACEHOLDER = fmt.Sprintf("v1.0.%s%d-FAKE", t.Format("20060102150405"), t.Nanosecond())
	editRef := stripSubDirFromTag(tag, mod.Module.Path)
	if usePlainReplace {
		editRef = PLACEHOLDER
	}
	requireArgs := []string{"mod", "edit", fmt.Sprintf("-require=%s@%s", mod.Module.Path, editRef)}
	fmt.Fprintf(os.Stderr, "go %s\n", strings.Join(requireArgs, " "))
	getCmd := exec.Command("go", requireArgs...)
	getCmd.Stderr = os.Stderr
	getCmd.Stdout = os.Stdout
	err = getCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to get module: %v", err)
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
