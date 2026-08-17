package vscodegit

import (
	"fmt"
	"os/exec"
	"strings"
)

var codeCommandOverride *string

// SetCodeCommandForTest overrides the code CLI path for tests. Pass empty string to reset.
func SetCodeCommandForTest(cmd string) {
	if cmd == "" {
		codeCommandOverride = nil
		return
	}
	s := cmd
	codeCommandOverride = &s
}

func resolveCodePath() (string, error) {
	if codeCommandOverride != nil {
		if *codeCommandOverride == "" {
			return "", fmt.Errorf("code: not found in PATH")
		}
		return *codeCommandOverride, nil
	}
	path, err := exec.LookPath("code")
	if err != nil {
		return "", fmt.Errorf("code: not found in PATH")
	}
	return path, nil
}

// EnsureCodeCLI verifies the VS Code `code` CLI is available.
func EnsureCodeCLI() error {
	_, err := resolveCodePath()
	return err
}

// EnsureExtensionListed verifies xhd2015.open-in-new-window is installed.
func EnsureExtensionListed() error {
	codePath, err := resolveCodePath()
	if err != nil {
		return err
	}
	cmd := exec.Command(codePath, "--list-extensions")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list VS Code extensions: %w", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == extensionID {
			return nil
		}
	}
	return fmt.Errorf(
		"%s is not installed\ninstall from the marketplace or run: code --install-extension %s",
		extensionID,
		extensionID,
	)
}

func runPrecheck() error {
	if err := EnsureCodeCLI(); err != nil {
		return err
	}
	return EnsureExtensionListed()
}
