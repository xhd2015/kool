package go_update

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// equivalent to go mod edit -replace "$mod=$absDir":
//
// dir=$1
// if [[ -z $dir ]];then
//     echo "requires dir" >&2
//     exit 1
// fi

// if [[ ! -d $dir ]];then
//     echo "no such dir: $dir" >&2
//     exit 1
// fi

// mod=$(cd "$dir" && go mod edit -json|jq -r '.Module.Path')
// if [[ -z $mod ]];then
//     echo "not a go module: $dir" >&2
//     exit 1
// fi
// absDir=$(cd "$dir" && pwd)
// go mod edit -replace "$mod=$absDir"

type ModuleInfo struct {
	Module struct {
		Path string `json:"Path"`
	} `json:"Module"`
}

// Replace checks if the given directory exists and contains a valid Go module,
// returns the absolute path of the directory and the module path
func Replace(dir string) (absDir string, modulePath string, err error) {
	// Check if directory is provided
	if dir == "" {
		return "", "", fmt.Errorf("requires dir")
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", "", fmt.Errorf("no such dir: %s", dir)
	}

	// Get absolute path
	absDir, err = filepath.Abs(dir)
	if err != nil {
		return "", "", fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Run 'go mod edit -json' in the directory
	cmd := exec.Command("go", "mod", "edit", "-json")
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("not a go module: %s", dir)
	}

	// Parse the JSON output
	var modInfo ModuleInfo
	if err := json.Unmarshal(output, &modInfo); err != nil {
		return "", "", fmt.Errorf("failed to parse module info: %v", err)
	}

	if modInfo.Module.Path == "" {
		return "", "", fmt.Errorf("not a go module: %s", dir)
	}
	// executing go mod edit -replace "$mod=$absDir"
	editCmd := exec.Command("go", "mod", "edit", "-replace", fmt.Sprintf("%s=%s", modInfo.Module.Path, absDir))
	editCmd.Stderr = os.Stderr
	editCmd.Stdout = os.Stdout
	err = editCmd.Run()
	if err != nil {
		return "", "", fmt.Errorf("failed to execute go mod edit -replace: %v", err)
	}

	return absDir, modInfo.Module.Path, nil
}
