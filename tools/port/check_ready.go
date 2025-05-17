package port

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/xhd2015/kool/pkgs/flag"
)

const help = `
Usage: check-port-ready [OPTIONS] <PORT>

Options:
  -t, --timeout <duration>  timeout for the check (default: 100ms), duration can be a number with unit (e.g. 1s, 100ms), or "never"
  --wait                    wait until port is ready

Examples:
  check-port-ready 8080
  check-port-ready --timeout 1s 8080
  check-port-ready --wait 8080
`

func CheckReady(args []string) error {
	const DEFAULT_NO_TIMEOUT = 100 * time.Millisecond
	var timeout time.Duration = DEFAULT_NO_TIMEOUT

	var verbose bool
	n := len(args)
	var remainArgs []string
	for i := 0; i < n; i++ {
		flag, value := flag.ParseFlag(args, &i)
		if flag == "" {
			remainArgs = append(remainArgs, args[i])
			continue
		}
		switch flag {
		case "-t", "--timeout":
			value, ok := value()
			if !ok {
				return fmt.Errorf("%s requires a value", flag)
			}
			if value == "never" {
				timeout = 0
			} else {
				var err error
				timeout, err = time.ParseDuration(value)
				if err != nil {
					return fmt.Errorf("invalid timeout: %w", err)
				}
			}
		case "-h", "--help":
			fmt.Print(strings.TrimPrefix(help, "\n"))
			return nil
		case "--wait":
			timeout = 0
		case "-v", "--verbose":
			verbose = true
		default:
			return fmt.Errorf("unknown flag: %s", flag)
		}
	}
	if len(remainArgs) == 0 {
		return fmt.Errorf("usage: check-port-ready <port>")
	}
	port, err := strconv.Atoi(remainArgs[0])
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}
	if port == 0 {
		return fmt.Errorf("port cannot be 0")
	}
	if port < 0 {
		return fmt.Errorf("port below 0 is invalid")
	}
	if port > 65535 {
		return fmt.Errorf("port above 65535 is invalid")
	}

	start := time.Now()
	// we can try dial or just lsof
	addr := fmt.Sprintf(":%d", port)

	for {
		// -P: not port names
		// -sTCP:LSTEN only show TCP connections with LISTEN state
		lsofCmd := exec.Command("lsof", "-P", "-sTCP:LISTEN", "-i", addr)
		output, err := lsofCmd.Output()
		if err != nil {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				return fmt.Errorf("error running lsof: %w", err)
			}
			// ignore exit code error
		}
		if strings.Contains(string(output), addr) {
			// ok
			if verbose {
				fmt.Printf("port %d is ready\n", port)
			}
			return nil
		}
		if timeout == DEFAULT_NO_TIMEOUT {
			return fmt.Errorf("port %d is not ready", port)
		}
		if timeout > 0 && time.Since(start) > timeout {
			return fmt.Errorf("port %d is not ready after %v", port, time.Since(start))
		}
		if verbose {
			fmt.Printf("port %d is not ready\n", port)
		}
		time.Sleep(1 * time.Second)
	}
}
