package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage:
  podman-machine-start [flags]

Flags:
  -v, --verbose     Verbose output
  -h, --help        Show help

Description:
  Checks if podman machine is running, and starts it if not.

Example:
  go run ./script/podman-machine-start
`

func main() {
	err := Handle(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func Handle(args []string) error {
	var verbose bool

	args, err := flags.
		Help("-h,--help", help).
		Bool("-v,--verbose", &verbose).
		Parse(args)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return fmt.Errorf("unexpected extra args: %v", args)
	}

	// Check if podman machine is running
	state, err := getPodmanMachineState()
	if err != nil {
		if verbose {
			fmt.Printf("Failed to get podman machine state: %v\n", err)
		}
		// If we can't get the state, try to start anyway
		return startPodmanMachine(verbose)
	}

	if verbose {
		fmt.Printf("Podman machine state: %s\n", state)
	}

	if state == "Running" {
		fmt.Println("Podman machine is already running")
		return nil
	}

	return startPodmanMachine(verbose)
}

func getPodmanMachineState() (string, error) {
	cmd := exec.Command("podman", "machine", "info", "--format", "{{.Host.MachineState}}")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}

func startPodmanMachine(verbose bool) error {
	fmt.Println("Starting podman machine...")

	cmd := exec.Command("podman", "machine", "start")
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start podman machine: %w", err)
	}

	fmt.Println("Podman machine started successfully")
	return nil
}
