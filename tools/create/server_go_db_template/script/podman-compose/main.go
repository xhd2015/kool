package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage:
  podman-compose [flags] [command]

Commands:
  up      Start containers (default)
  down    Stop and remove containers

Flags:
  --build           Build images before starting
  -d, --detach      Run containers in background
  -v, --verbose     Verbose output
  -h, --help        Show help

Example:
  go run ./script/podman-compose
  go run ./script/podman-compose up --build
  go run ./script/podman-compose down
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
	var build bool
	var detach bool

	args, err := flags.
		Help("-h,--help", help).
		Bool("-v,--verbose", &verbose).
		Bool("--build", &build).
		Bool("-d,--detach", &detach).
		Parse(args)
	if err != nil {
		return err
	}

	command := "up"
	if len(args) > 0 {
		command = args[0]
	}

	// Check if podman machine is running
	if err := checkPodmanMachine(); err != nil {
		return err
	}

	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	// Use the compose file from script/podman-compose directory
	composeFile := filepath.Join(projectRoot, "script", "podman-compose", "podman-compose.yml")

	// Ensure mysql_db directory exists
	mysqlDBDir := filepath.Join(projectRoot, "mysql_db")
	if err := os.MkdirAll(mysqlDBDir, 0755); err != nil {
		return fmt.Errorf("failed to create mysql_db directory: %w", err)
	}

	// Set environment variables
	env := os.Environ()
	env = append(env, "PROJECT_ROOT="+projectRoot)
	env = append(env, getMySQLEnv()...)

	// Build command
	cmdArgs := []string{"-f", composeFile, command}
	if command == "up" {
		if build {
			cmdArgs = append(cmdArgs, "--build")
		}
		if detach {
			cmdArgs = append(cmdArgs, "-d")
		}
	}

	if verbose {
		fmt.Printf("Running: podman-compose %v\n", cmdArgs)
	}

	cmd := exec.Command("podman-compose", cmdArgs...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = projectRoot

	return cmd.Run()
}

// checkPodmanMachine checks if the podman machine is running
func checkPodmanMachine() error {
	cmd := exec.Command("podman", "machine", "info", "--format", "{{.Host.MachineState}}")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("podman machine is not available, please run: podman machine start")
	}

	state := strings.TrimSpace(stdout.String())
	if state != "Running" {
		return fmt.Errorf("podman machine is not running (state: %s), please run: podman machine start", state)
	}

	return nil
}

func getMySQLEnv() []string {
	var env []string

	// Always use MySQL 8.0
	env = append(env, "MYSQL_VERSION=8.0")

	// Set user mapping
	uid := os.Getuid()
	gid := os.Getgid()
	env = append(env, "MYSQL_USER_MAPPING="+strconv.Itoa(uid)+":"+strconv.Itoa(gid))

	return env
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}
