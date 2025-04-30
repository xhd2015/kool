package resolve

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type ModuleInfo struct {
	Module struct {
		Path string `json:"Path"`
	} `json:"Module"`
	Require []struct {
		Path    string `json:"Path"`
		Version string `json:"Version"`
	} `json:"Require"`
}

func GetModuleInfo(dir string) (*ModuleInfo, error) {
	// Run 'go mod edit -json' in the directory
	cmd := exec.Command("go", "mod", "edit", "-json")
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("resolve go mod: %s %w", dir, err)
	}

	// Parse the JSON output
	var modInfo ModuleInfo
	if err := json.Unmarshal(output, &modInfo); err != nil {
		return nil, fmt.Errorf("failed to parse module info: %v", err)
	}

	return &modInfo, nil
}
