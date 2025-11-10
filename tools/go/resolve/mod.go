package resolve

import (
	"encoding/json"
	"fmt"

	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/git"
)

type ModuleInfo struct {
	Module struct {
		Path string `json:"Path"`
	} `json:"Module"`
	Require []struct {
		Path    string `json:"Path"`
		Version string `json:"Version"`
	} `json:"Require"`
	Replace []struct {
		Old struct {
			Path string `json:"Path"`
		} `json:"Old"`
		New struct {
			Path    string `json:"Path"`
			Version string `json:"Version"`
		} `json:"New"`
	} `json:"Replace"`
}

func GetModuleInfo(dir string) (*ModuleInfo, error) {
	// Run 'go mod edit -json' in the directory
	output, err := cmd.Dir(dir).Output("go", "mod", "edit", "-json")
	if err != nil {
		return nil, fmt.Errorf("resolve go mod: %s %w", dir, err)
	}

	// Parse the JSON output
	var modInfo ModuleInfo
	if err := json.Unmarshal([]byte(output), &modInfo); err != nil {
		return nil, fmt.Errorf("failed to parse module info: %v", err)
	}

	return &modInfo, nil
}

// GetRootModulePath gets the module path from the go.mod at the root of the git repository
func GetRootModulePath(targetDir string) (string, error) {
	// Find the git root
	gitRoot, err := git.ShowTopLevel(targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to get git root for %s: %w", targetDir, err)
	}

	// Get the module info from the git root
	rootModInfo, err := GetModuleInfo(gitRoot)
	if err != nil {
		return "", fmt.Errorf("failed to get root module info for %s: %w", gitRoot, err)
	}

	return rootModInfo.Module.Path, nil
}
