//go:build ignore

package server

import (
	"os"
	"path/filepath"
)

// ResolveStateDir resolves the daemon state directory from flag or environment.
func ResolveStateDir(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if env := os.Getenv("DAEMON_STATE_DIR"); env != "" {
		return env, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "__STATE_SUBPATH__"), nil
}