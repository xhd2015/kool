package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/xhd2015/less-flags"
)

const help = `
Usage: dev [__PROJECT_NAME__ options]

Start __PROJECT_NAME__ in dev mode.
This wrapper runs:

  go run ./ --dev [__PROJECT_NAME__ options...]

Examples:
  go run ./script/dev
  go run ./script/dev --port 9000
  go run ./script/dev --route-prefix __PROJECT_NAME__
  go run ./script/dev -- --help
`

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	forwardArgs, err := lessflags.Help("-h,--help", help).Parse(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	root, err := moduleRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve module root: %v\n", err)
		return 1
	}

	forwardArgs = stripLeadingDoubleDash(forwardArgs)
	cmdArgs := append([]string{"run", "./", "--dev"}, forwardArgs...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = root
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	fmt.Printf("+ cd %s\n", root)
	fmt.Printf("+ go %s\n", strings.Join(cmdArgs, " "))

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start dev server: %v\n", err)
		return 1
	}

	go func() {
		sig := <-sigCh
		if sig == nil || cmd.Process == nil {
			return
		}
		if signalValue, ok := sig.(syscall.Signal); ok {
			_ = syscall.Kill(-cmd.Process.Pid, signalValue)
			return
		}
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}()

	if err := cmd.Wait(); err != nil {
		return exitCode(err)
	}
	return 0
}

func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func stripLeadingDoubleDash(args []string) []string {
	if len(args) > 0 && args[0] == "--" {
		return args[1:]
	}
	return args
}

func exitCode(err error) int {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		fmt.Fprintf(os.Stderr, "dev server failed: %v\n", err)
		return 1
	}
	if status, ok := exitErr.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		return 128 + int(status.Signal())
	}
	return exitErr.ExitCode()
}
