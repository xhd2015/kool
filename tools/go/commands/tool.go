package commands

import (
	"fmt"
	"os/exec"
	"strings"
)

// GoToolNM executes "go tool nm" on the specified binary.
func GoToolNM(binary string) (string, error) {
	output, err := exec.Command("go", "tool", "nm", binary).Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute go tool nm %s: %w", binary, err)
	}
	return string(output), nil
}

// GoToolObjdump executes "go tool objdump" with the specified arguments.
func GoToolObjdump(args ...string) (string, error) {
	cmdArgs := append([]string{"tool", "objdump"}, args...)
	output, err := exec.Command("go", cmdArgs...).Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute go tool objdump %s: %w", strings.Join(args, " "), err)
	}
	return string(output), nil
}