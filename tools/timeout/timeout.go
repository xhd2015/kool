package timeout

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
kool timeout - Run a command with a timeout

Usage: kool timeout <duration> <command> [args...]

Arguments:
  duration                         timeout duration (e.g., 1s, 30s, 1m, 1h)
  command                          command to run
  args                             arguments for the command

Options:
  -h,--help                        show help message

Examples:
  kool timeout 5s sleep 10         run 'sleep 10' with 5 second timeout
  kool timeout 1m curl example.com run 'curl example.com' with 1 minute timeout
  kool timeout 30s go test ./...   run 'go test ./...' with 30 second timeout
`

func Handle(args []string) error {
	args, err := flags.Help("-h,--help", help).StopOnFirstArg().Parse(args)
	if err != nil {
		return err
	}

	if len(args) < 2 {
		return fmt.Errorf("requires at least duration and command: kool timeout <duration> <command> [args...]")
	}

	durationStr := args[0]
	command := args[1]
	cmdArgs := args[2:]

	// Parse duration
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("invalid duration '%s': %v", durationStr, err)
	}

	if duration <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}

	return runWithTimeout(duration, command, cmdArgs)
}

func runWithTimeout(timeout time.Duration, command string, args []string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create the command with context
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start the command
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start command '%s': %v", command, err)
	}

	// Wait for command completion or signals
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		// Command completed
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				// Command exited with non-zero status
				return &ExitCodeError{ExitCode: exitError.ExitCode()}
			}
			return fmt.Errorf("command failed: %v", err)
		}
		return nil

	case <-ctx.Done():
		// Timeout occurred
		if cmd.Process != nil {
			// Try graceful termination first
			cmd.Process.Signal(os.Interrupt)

			// Give it a moment to terminate gracefully
			select {
			case <-done:
				return &TimeoutError{Duration: timeout, Command: formatCommand(command, args)}
			case <-time.After(2 * time.Second):
				// Force kill if it doesn't terminate gracefully
				cmd.Process.Kill()
				return &TimeoutError{Duration: timeout, Command: formatCommand(command, args)}
			}
		}
		return &TimeoutError{Duration: timeout, Command: formatCommand(command, args)}

	case sig := <-sigChan:
		// Received interrupt signal
		if cmd.Process != nil {
			cmd.Process.Signal(sig)
		}
		return &InterruptError{Signal: sig}
	}
}

func formatCommand(command string, args []string) string {
	if len(args) == 0 {
		return command
	}
	return command + " " + strings.Join(args, " ")
}

// Custom error types for different exit scenarios

type ExitCodeError struct {
	ExitCode int
}

func (e *ExitCodeError) Error() string {
	return fmt.Sprintf("command exited with code %d", e.ExitCode)
}

func (e *ExitCodeError) SilenceExitCode() int {
	return e.ExitCode
}

type TimeoutError struct {
	Duration time.Duration
	Command  string
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("command '%s' timed out after %v", e.Command, e.Duration)
}

func (e *TimeoutError) SilenceExitCode() int {
	return 124 // Standard timeout exit code
}

type InterruptError struct {
	Signal os.Signal
}

func (e *InterruptError) Error() string {
	return fmt.Sprintf("command interrupted by signal %v", e.Signal)
}

func (e *InterruptError) SilenceExitCode() int {
	return 130 // Standard interrupt exit code
}
