package go_update

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/xhd2015/kool/tools/git_tag_next"
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
	if tag == "" {
		tag, _ = git_tag_next.ShowCurrentBranch(dir)
	}
	if tag == "" {
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
	requireArgs := []string{"mod", "edit", fmt.Sprintf("-require=%s@%s", mod.Module.Path, tag)}
	fmt.Fprintf(os.Stderr, "go %s\n", strings.Join(requireArgs, " "))
	getCmd := exec.Command("go", requireArgs...)
	getCmd.Stderr = os.Stderr
	getCmd.Stdout = os.Stdout
	err = getCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to get module: %v", err)
	}

	msgCmd := exec.Command("git", "log", "-1", "--format=%s", tag)
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
