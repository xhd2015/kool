package go_update

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
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

	tagCmd := exec.Command("git", "tag", "-l", "--points-at", "HEAD")
	tagCmd.Dir = dir
	tagCmd.Stderr = os.Stderr
	tagOutput, err := tagCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get tag: %v", err)
	}
	tag := strings.TrimSpace(string(tagOutput))
	if tag == "" {
		return fmt.Errorf("no tag at HEAD: %s", dir)
	}

	editCmd := exec.Command("go", "mod", "edit", "-dropreplace", mod.Module.Path)
	editCmd.Dir = dir
	editCmd.Stderr = os.Stderr
	editCmd.Stdout = os.Stdout
	err = editCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to edit go mod: %v", err)
	}

	getCmd := exec.Command("go", "get", fmt.Sprintf("%s@%s", mod.Module.Path, tag))
	getCmd.Dir = dir
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
