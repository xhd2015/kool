//go:build ignore

package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/dev/hot_reload/air"
	"__MODULE_NAME__/server"
	"github.com/xhd2015/kool/pkgs/web"
	"github.com/xhd2015/less-flags"
	"github.com/xhd2015/xgo/support/cmd"
)

const help = `
Usage: go run ./script/dev [options]

Dev server: Go API (air) + Vite frontend (HMR).
Vite listens on :6193 (strictPort). Air rebuilds the Go API on .go changes;
Vite stays up. Prefer this over plain 'go run ./cmd/__PROJECT_NAME__ --dev'.

Options:
  --port <n>        Go listen port (default: first free from 8080)
  --vite-port <n>   Vite port (default: 6193)
  --no-air          Run Go once without air (still uses external Vite)
  --no-open         Do not open browser
  -h, --help        Show help
`

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if err := handle(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func handle(args []string) error {
	var portFlag int
	vitePort := server.DefaultVitePort
	var noAir bool
	var noOpen bool
	remain, err := lessflags.
		Int("--port", &portFlag).
		Int("--vite-port", &vitePort).
		Bool("--no-air", &noAir).
		Bool("--no-open", &noOpen).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remain) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(remain, " "))
	}
	if vitePort <= 0 {
		vitePort = server.DefaultVitePort
	}

	root, err := moduleRoot()
	if err != nil {
		return err
	}
	reactDir := filepath.Join(root, "__PROJECT_NAME__-react")

	if _, err := os.Stat(filepath.Join(reactDir, "node_modules")); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "installing frontend deps (bun install)...")
		if err := cmd.Debug().Dir(reactDir).Run("bun", "install"); err != nil {
			return fmt.Errorf("bun install: %w", err)
		}
	}

	if portListening(vitePort) {
		pids, err := listenPIDs(vitePort)
		if err != nil {
			return fmt.Errorf("inspect vite port %d: %w", vitePort, err)
		}
		fmt.Fprintf(os.Stderr, "warning: port %d busy (pid %v); stopping it\n", vitePort, pids)
		if err := killListenersOnPort(vitePort); err != nil {
			return err
		}
	}

	viteCmd := exec.Command("bun", "run", "dev", "--", "--port", strconv.Itoa(vitePort), "--strictPort")
	viteCmd.Dir = reactDir
	viteCmd.Stdout = os.Stdout
	viteCmd.Stderr = os.Stderr
	viteCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	fmt.Fprintf(os.Stderr, "starting vite on :%d...\n", vitePort)
	if err := viteCmd.Start(); err != nil {
		return fmt.Errorf("start vite: %w", err)
	}

	viteExited := make(chan error, 1)
	go func() {
		viteExited <- viteCmd.Wait()
	}()

	var stopViteOnce sync.Once
	stopVite := func() {
		stopViteOnce.Do(func() {
			fmt.Fprintln(os.Stderr, "stopping vite...")
			if viteCmd.Process != nil {
				pgid := viteCmd.Process.Pid
				_ = syscall.Kill(-pgid, syscall.SIGTERM)
				select {
				case <-viteExited:
				case <-time.After(3 * time.Second):
					_ = syscall.Kill(-pgid, syscall.SIGKILL)
					select {
					case <-viteExited:
					case <-time.After(2 * time.Second):
					}
				}
			}
			if portListening(vitePort) {
				_ = killListenersOnPort(vitePort)
			}
		})
	}
	defer stopVite()

	fmt.Fprint(os.Stderr, "waiting for vite")
	deadline := time.Now().Add(45 * time.Second)
	ready := false
	for time.Now().Before(deadline) {
		select {
		case err := <-viteExited:
			fmt.Fprintln(os.Stderr)
			if err != nil {
				return fmt.Errorf("vite exited before becoming ready: %w", err)
			}
			return fmt.Errorf("vite exited before becoming ready")
		default:
		}
		if portListening(vitePort) {
			fmt.Fprintln(os.Stderr, " ready!")
			ready = true
			break
		}
		time.Sleep(250 * time.Millisecond)
		fmt.Fprint(os.Stderr, ".")
	}
	if !ready {
		fmt.Fprintln(os.Stderr)
		return fmt.Errorf("vite :%d failed to start within timeout", vitePort)
	}

	if noAir {
		return runWithoutAir(root, portFlag, vitePort, noOpen, stopVite, viteExited)
	}
	return runWithAir(root, portFlag, vitePort, noOpen, stopVite, viteExited)
}

func serverBinArgs(portFlag, vitePort int) []string {
	args := []string{
		"--dev",
		"--external-vite",
		"--vite-port", strconv.Itoa(vitePort),
	}
	if portFlag != 0 {
		args = append(args, "--port", strconv.Itoa(portFlag))
	}
	return args
}

func runWithoutAir(root string, portFlag, vitePort int, noOpen bool, stopVite func(), viteExited <-chan error) error {
	listenPort := portFlag
	if listenPort == 0 {
		var err error
		listenPort, err = web.FindAvailablePort(8080, 100)
		if err != nil {
			return err
		}
	}

	binArgs := serverBinArgs(listenPort, vitePort)
	cmdArgs := append([]string{"run", "./cmd/__PROJECT_NAME__"}, binArgs...)
	goCmd := exec.Command("go", cmdArgs...)
	goCmd.Dir = root
	goCmd.Stdout = os.Stdout
	goCmd.Stderr = os.Stderr
	goCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	fmt.Fprintf(os.Stderr, "starting go (no air) → :%d ...\n", listenPort)
	fmt.Fprintf(os.Stderr, "  go %s\n", strings.Join(cmdArgs, " "))
	if err := goCmd.Start(); err != nil {
		return fmt.Errorf("start go: %w", err)
	}

	if !noOpen {
		go openWhenReady(listenPort)
	}

	goExited := make(chan error, 1)
	go func() {
		goExited <- goCmd.Wait()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigCh:
		if goCmd.Process != nil {
			_ = syscall.Kill(-goCmd.Process.Pid, syscall.SIGTERM)
		}
		stopVite()
		return nil
	case err := <-viteExited:
		if goCmd.Process != nil {
			_ = syscall.Kill(-goCmd.Process.Pid, syscall.SIGTERM)
		}
		if err != nil {
			return fmt.Errorf("vite exited: %w", err)
		}
		return fmt.Errorf("vite exited")
	case err := <-goExited:
		stopVite()
		if err != nil {
			return fmt.Errorf("go exited: %w", err)
		}
		return fmt.Errorf("go exited")
	}
}

func runWithAir(root string, portFlag, vitePort int, noOpen bool, stopVite func(), viteExited <-chan error) error {
	listenPort := portFlag
	if listenPort == 0 {
		var err error
		listenPort, err = web.FindAvailablePort(8080, 100)
		if err != nil {
			return err
		}
	}

	tmpDir := filepath.Join(root, "tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("mkdir tmp: %w", err)
	}
	binPath := filepath.Join(tmpDir, "__PROJECT_NAME__-dev")
	relBin := "./tmp/__PROJECT_NAME__-dev"
	buildCmd := "go build -o ./tmp/__PROJECT_NAME__-dev ./cmd/__PROJECT_NAME__"
	binArgs := serverBinArgs(listenPort, vitePort)

	fmt.Fprintf(os.Stderr, "starting air (BE watch) → :%d ...\n", listenPort)
	proc, err := air.Start(context.Background(), air.Options{
		Dir:        root,
		BuildCmd:   buildCmd,
		Entrypoint: relBin,
		ArgsBin:    binArgs,
		ExcludeDir: []string{"tmp", "__PROJECT_NAME__-react", "node_modules", ".git", "vendor", "script"},
		OnStop: func() {
			if portListening(listenPort) {
				_ = killListenersOnPort(listenPort)
			}
			_ = os.Remove(binPath)
		},
	})
	if err != nil {
		return err
	}
	defer proc.Stop()

	if !noOpen {
		go openWhenReady(listenPort)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigCh:
		proc.Stop()
		stopVite()
		return nil
	case err := <-viteExited:
		proc.Stop()
		if err != nil {
			return fmt.Errorf("vite exited: %w", err)
		}
		return fmt.Errorf("vite exited")
	case err := <-proc.Exited:
		stopVite()
		if err != nil {
			return fmt.Errorf("air exited: %w", err)
		}
		return fmt.Errorf("air exited")
	}
}

func openWhenReady(listenPort int) {
	url := fmt.Sprintf("http://127.0.0.1:%d/", listenPort)
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if portListening(listenPort) {
			time.Sleep(800 * time.Millisecond)
			_ = web.OpenBrowser(url)
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "__PROJECT_NAME__-react")); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod + __PROJECT_NAME__-react not found")
		}
		dir = parent
	}
}

func portListening(port int) bool {
	for _, host := range []string{"localhost", "127.0.0.1", "[::1]"} {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 200*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return true
		}
	}
	return false
}

func listenPIDs(port int) ([]int, error) {
	out, err := exec.Command("lsof", "-nP", fmt.Sprintf("-iTCP:%d", port), "-sTCP:LISTEN", "-t").Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, err
	}
	var pids []int
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

func killListenersOnPort(port int) error {
	pids, err := listenPIDs(port)
	if err != nil {
		return err
	}
	for _, pid := range pids {
		_ = syscall.Kill(pid, syscall.SIGTERM)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if !portListening(port) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	pids, err = listenPIDs(port)
	if err != nil {
		return err
	}
	for _, pid := range pids {
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
	time.Sleep(200 * time.Millisecond)
	if portListening(port) {
		return fmt.Errorf("port %d still busy after kill", port)
	}
	return nil
}
