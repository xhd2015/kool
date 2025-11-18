package update

import (
	"fmt"
	"os"

	"github.com/xhd2015/kool/tools/go/resolve"
)

func UpdateReplaced(dir string) error {
	// Get current directory's go.mod info
	modInfo, err := resolve.GetModuleInfo(dir)
	if err != nil {
		return fmt.Errorf("failed to get module info: %w", err)
	}

	if len(modInfo.Replace) == 0 {
		fmt.Fprintf(os.Stderr, "no replacements found\n")
		return nil
	}

	// Use the shared function to collect replacement update infos
	updateInfos, err := collectReplacementUpdateInfos(modInfo)
	if err != nil {
		return fmt.Errorf("failed to collect replacement info: %w", err)
	}

	if len(updateInfos) == 0 {
		fmt.Fprintf(os.Stderr, "no local replacements to update\n")
		return nil
	}

	// Execute updates using the shared function
	if err := executeModuleUpdates(dir, updateInfos); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "successfully updated %d replacement(s)\n", len(updateInfos))
	return nil
}
