// Package ssh provides kool ssh utilities (local port forward, etc.).
package ssh

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/xhd2015/kool/pkgs/flag"
)

const rootHelp = `kool ssh - SSH helpers

Usage:
  kool ssh forward --local PORT --to-remote-internal HOST:PORT --host SSH_HOST
  kool ssh -h|--help

Commands:
  forward   local port forward (-L); blocks until exit (Ctrl+C cleans up)

Examples:
  kool ssh forward --local 18082 --to-remote-internal 127.0.0.1:8082 --host test.devbox
`

const forwardHelp = `kool ssh forward - local SSH port forward (-L)

Forwards a local port to an address reachable only on the remote host
(remote-internal). Blocks in the foreground; Ctrl+C or process exit stops
ssh and clears the forward (no background daemon).

Usage:
  kool ssh forward --local PORT --to-remote-internal HOST:PORT --host SSH_HOST

Options:
  --local PORT                      local listen port (bound on 127.0.0.1)
  --to-remote-internal HOST:PORT    target as seen on the remote machine
                                    (e.g. 127.0.0.1:8082)
  --host SSH_HOST                   SSH host (e.g. test.devbox from ~/.ssh/config)
  --no-check                        skip local readiness probe after connect
  -h,--help                         show help message

Examples:
  kool ssh forward --local 18082 --to-remote-internal 127.0.0.1:8082 --host test.devbox
`

// Handle is the production entry for kool ssh.
func Handle(args []string) error {
	return HandleWith(args, HandleOpts{})
}

// HandleOpts injects IO and process runner for tests.
type HandleOpts struct {
	Stdout io.Writer
	Stderr io.Writer
	// RunSSH runs the ssh process and blocks until exit. Nil → real exec.
	// argv is the full argument list after "ssh" (not including binary name).
	RunSSH func(argv []string) error
	// WaitLocalReady waits until localPort accepts TCP. Nil → default dial loop.
	WaitLocalReady func(localPort int, timeout time.Duration) error
}

// HandleWith is the injectable entry used by tests.
func HandleWith(args []string, opts HandleOpts) error {
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	if len(args) == 0 {
		return fmt.Errorf("requires subcommand, try 'kool ssh --help'")
	}

	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Fprint(stdout, rootHelp)
		if !strings.HasSuffix(rootHelp, "\n") {
			fmt.Fprintln(stdout)
		}
		return nil
	}

	switch args[0] {
	case "forward":
		return handleForward(args[1:], opts, stdout)
	default:
		return fmt.Errorf("unrecognized command: %s", args[0])
	}
}

func handleForward(args []string, opts HandleOpts, stdout io.Writer) error {
	var localStr, toRemote, host string
	var localSet, toSet, hostSet bool
	var noCheck bool

	n := len(args)
	for i := 0; i < n; i++ {
		f, value := flag.ParseFlag(args, &i)
		if f == "" {
			return fmt.Errorf("unexpected argument: %s", args[i])
		}
		switch f {
		case "-h", "--help":
			fmt.Fprint(stdout, forwardHelp)
			if !strings.HasSuffix(forwardHelp, "\n") {
				fmt.Fprintln(stdout)
			}
			return nil
		case "--local":
			v, ok := value()
			if !ok {
				return fmt.Errorf("--local requires a value")
			}
			localStr = v
			localSet = true
		case "--to-remote-internal":
			v, ok := value()
			if !ok {
				return fmt.Errorf("--to-remote-internal requires a value")
			}
			toRemote = v
			toSet = true
		case "--host":
			v, ok := value()
			if !ok {
				return fmt.Errorf("--host requires a value")
			}
			host = v
			hostSet = true
		case "--no-check":
			noCheck = true
		default:
			return fmt.Errorf("unrecognized: %s", f)
		}
	}

	if !localSet || strings.TrimSpace(localStr) == "" {
		return fmt.Errorf("requires --local")
	}
	if !toSet || strings.TrimSpace(toRemote) == "" {
		return fmt.Errorf("requires --to-remote-internal")
	}
	if !hostSet || strings.TrimSpace(host) == "" {
		return fmt.Errorf("requires --host")
	}

	localPort, err := parseLocalPort(localStr)
	if err != nil {
		return err
	}
	remoteHost, remotePort, err := parseHostPort(toRemote)
	if err != nil {
		return fmt.Errorf("--to-remote-internal: %w", err)
	}

	argv := BuildSSHForwardArgv(localPort, remoteHost, remotePort, host)
	cmdLine := formatSSHCommand(argv)

	fmt.Fprintf(stdout, "Forwarding local :%d → %s (remote-internal %s:%d)\n", localPort, host, remoteHost, remotePort)
	fmt.Fprintf(stdout, "Local URL: http://127.0.0.1:%d\n", localPort)
	fmt.Fprintln(stdout, cmdLine)
	fmt.Fprintln(stdout, "Press Ctrl+C to stop")

	// Injected runner: tests supply full lifecycle (no real network).
	if opts.RunSSH != nil {
		return opts.RunSSH(argv)
	}

	cmd := exec.Command("ssh", argv...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start ssh: %w", err)
	}

	if !noCheck {
		wait := opts.WaitLocalReady
		if wait == nil {
			wait = waitLocalPort
		}
		if err := wait(localPort, 8*time.Second); err != nil {
			_ = cmd.Process.Kill()
			_, _ = cmd.Process.Wait()
			return fmt.Errorf("local port %d not ready after forward: %w", localPort, err)
		}
	}

	err = cmd.Wait()
	if err != nil {
		// Non-zero exit from ssh (including signal) is reported as-is.
		return err
	}
	return nil
}

// BuildSSHForwardArgv returns ssh args after the binary name for a local forward.
func BuildSSHForwardArgv(localPort int, remoteHost string, remotePort int, sshHost string) []string {
	forward := fmt.Sprintf("%d:%s:%d", localPort, remoteHost, remotePort)
	return []string{
		"-N",
		"-o", "ExitOnForwardFailure=yes",
		"-o", "ServerAliveInterval=30",
		"-L", forward,
		sshHost,
	}
}

func formatSSHCommand(argv []string) string {
	parts := make([]string, 0, len(argv)+1)
	parts = append(parts, "ssh")
	for _, a := range argv {
		if strings.ContainsAny(a, " \t\"'") {
			parts = append(parts, strconv.Quote(a))
		} else {
			parts = append(parts, a)
		}
	}
	return strings.Join(parts, " ")
}

func parseLocalPort(s string) (int, error) {
	s = strings.TrimSpace(s)
	// allow optional 127.0.0.1:PORT
	if i := strings.LastIndex(s, ":"); i >= 0 {
		s = s[i+1:]
	}
	p, err := strconv.Atoi(s)
	if err != nil || p < 1 || p > 65535 {
		return 0, fmt.Errorf("invalid --local port: %s", s)
	}
	return p, nil
}

func parseHostPort(s string) (host string, port int, err error) {
	s = strings.TrimSpace(s)
	host, portStr, ok := strings.Cut(s, ":")
	if !ok || host == "" || portStr == "" {
		return "", 0, fmt.Errorf("want HOST:PORT, got %q", s)
	}
	port, err = strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return "", 0, fmt.Errorf("invalid port in %q", s)
	}
	return host, port, nil
}

func waitLocalPort(localPort int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	addr := fmt.Sprintf("127.0.0.1:%d", localPort)
	var last error
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		last = err
		time.Sleep(100 * time.Millisecond)
	}
	if last == nil {
		last = fmt.Errorf("timeout")
	}
	return last
}
