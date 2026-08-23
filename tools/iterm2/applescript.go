package iterm2

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// defaultRunAppleScript runs an AppleScript body via osascript (restore / live_scan).
// Snapshot hierarchy collection lives in shell/iterm2/snapshot — not here.
func defaultRunAppleScript(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s", msg)
	}
	return stdout.String(), nil
}
