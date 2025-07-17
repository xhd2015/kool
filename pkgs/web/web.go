package web

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
)

// FindAvailablePort finds an available port starting from the given port.
// Returns the available port or an error if none found within maxAttempts.
func FindAvailablePort(startPort int, maxAttempts int) (int, error) {
	for i := 0; i < maxAttempts; i++ {
		currentPort := startPort + i
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", currentPort))
		if err == nil {
			listener.Close()
			return currentPort, nil
		}
	}
	return -1, fmt.Errorf("could not find available port after %d attempts", maxAttempts)
}

// OpenBrowser opens a URL in the default browser.
func OpenBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open browser: %v", err)
	}

	return nil
}
