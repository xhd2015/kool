package port

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func HandleKill(args []string) error {
	// lsof -iTCP:15000 -sTCP:LISTEN -t
	//   -iTCP:15000: only TCP listen on port 15000
	//   -sTCP:LISTEN: only show listening socket
	//   -t: only show pid
	if len(args) == 0 {
		return fmt.Errorf("usage: kool kill-port <port>")
	}
	port, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("port: %w", err)
	}
	args = args[1:]
	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra argument: %s", strings.Join(args, " "))
	}
	pidOutput, err := exec.Command("lsof", "-iTCP:"+strconv.Itoa(port), "-sTCP:LISTEN", "-t").Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			fmt.Fprintf(os.Stderr, "no process on port %d\n", port)
			return nil
		}
		return err
	}
	pid := strings.TrimSpace(string(pidOutput))
	if pid == "" {
		fmt.Fprintf(os.Stderr, "no process on port %d\n", port)
		return nil
	}
	fmt.Printf("kill -9 %s\n", pid)
	killCmd := exec.Command("kill", "-9", pid)
	killCmd.Stdout = os.Stdout
	killCmd.Stderr = os.Stderr
	return killCmd.Run()
}
