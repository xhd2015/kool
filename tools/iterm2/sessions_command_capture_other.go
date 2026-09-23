//go:build !darwin

package iterm2

import "fmt"

func readForegroundIdentity(pid int) (foregroundProcessIdentity, error) {
	return foregroundProcessIdentity{}, fmt.Errorf("foreground process identity is unsupported on this platform")
}

func readForegroundArgv(pid int) ([]string, error) {
	return nil, fmt.Errorf("exact foreground argv is unsupported on this platform")
}
func readForegroundCwd(pid int) (string, error) {
	return "", fmt.Errorf("foreground cwd is unsupported on this platform")
}
