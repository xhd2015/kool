package replace

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/xhd2015/kool/tools/go/resolve"
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

	modInfo, err := resolve.GetModuleInfo(dir)
	if err != nil {
		return "", "", err
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
