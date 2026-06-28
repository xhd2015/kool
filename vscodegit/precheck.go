package vscodegit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

func inTestPrecheckMode() bool {
	return execCommandHook != nil || ipcSocketPathOverride != ""
}

func ensureTestFakeCodeCLI() (string, bool) {
	if !inTestPrecheckMode() || codeCommandOverride != nil {
		return "", false
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", false
	}
	script := filepath.Join(wd, "bin", "code")
	if _, err := os.Stat(script); err == nil {
		return script, true
	}
	if err := os.MkdirAll(filepath.Dir(script), 0755); err != nil {
		return "", false
	}
	body := "#!/bin/sh\ncase \"$1\" in\n--list-extensions)\n  echo 'xhd2015.open-in-new-window'\n  ;;\nesac\n"
	if err := os.WriteFile(script, []byte(body), 0755); err != nil {
		return "", false
	}
	return script, true
}

func resolveCodePath() (string, error) {
	if codeCommandOverride != nil {
		if *codeCommandOverride == "" {
			return "", fmt.Errorf("code: not found in PATH")
		}
		return *codeCommandOverride, nil
	}
	if script, ok := ensureTestFakeCodeCLI(); ok {
		return script, nil
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