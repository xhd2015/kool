package for_every

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/xhd2015/kool/pkgs/duration"
	"github.com/xhd2015/kool/pkgs/errs"
	"github.com/xhd2015/less-flags"
)

const help = `
kool for-every - Run a command repeatedly with a sleep between iterations

Usage:
  kool for-every [OPTIONS] <duration> <command> [args...]
  kool for-every-<duration> [OPTIONS] <command> [args...]

Arguments:
  duration                         interval between runs (e.g. 60s, 1m, 500ms, or bare int seconds)
  command                          command to run each iteration
  args                             arguments for the command

Options:
  --max-runs N                     stop after N iterations (N > 0; omit = infinite)
  --max-failure N                  exit after N consecutive command failures (N > 0; 0/omit = unlimited)
  --allow-failure                  exit on first command failure (≡ --max-failure 1 when --max-failure unset;
                                   if both set, --max-failure wins)
  -h,--help                        show help message

Notes:
  - First run is immediate (no initial sleep); sleep happens after each run.
  - Default on failure: log to stderr and continue; a success resets the consecutive-failure counter.
  - Child inherits stdin/stdout/stderr.

Examples:
  kool for-every 60s echo tick
  kool for-every-60s echo tick
  kool for-every --max-runs 3 10ms true
  kool for-every --allow-failure --max-runs 10 1s make test
  kool for-every --max-failure 2 5s curl -f localhost:8080/health
`

// Handle implements spaced form: kool for-every [flags] <duration> <command> [args...]
func Handle(args []string) error {
	return handle(false, "", args)
}

// HandleWithDuration implements glued form: kool for-every-<duration> [flags] <command> [args...]
func HandleWithDuration(durationStr string, args []string) error {
	return handle(true, durationStr, args)
}

func handle(glued bool, gluedDuration string, args []string) error {
	maxRuns := -1 // -1 = unset (infinite); when set must be > 0
	maxFailure := 0
	allowFailure := false

	remain, err := lessflags.
		Int("--max-runs", &maxRuns).
		Int("--max-failure", &maxFailure).
		Bool("--allow-failure", &allowFailure).
		Help("-h,--help", help).
		StopOnFirstArg().
		Parse(args)
	if err != nil {
		return err
	}

	if maxRuns == 0 || maxRuns < -1 {
		return fmt.Errorf("--max-runs must be greater than 0")
	}
	if maxFailure < 0 {
		return fmt.Errorf("--max-failure must be >= 0")
	}

	var durationStr string
	var cmdAndArgs []string
	if glued {
		durationStr = gluedDuration
		cmdAndArgs = remain
	} else {
		if len(remain) == 0 {
			return fmt.Errorf("requires duration and command: kool for-every [OPTIONS] <duration> <command> [args...]")
		}
		durationStr = remain[0]
		cmdAndArgs = remain[1:]
	}

	if durationStr == "" {
		return fmt.Errorf("requires duration: kool for-every [OPTIONS] <duration> <command> [args...]")
	}

	interval, err := duration.Parse(durationStr)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %v", durationStr, err)
	}

	if len(cmdAndArgs) == 0 {
		return fmt.Errorf("requires command: kool for-every [OPTIONS] <duration> <command> [args...]")
	}

	command := cmdAndArgs[0]
	cmdArgs := cmdAndArgs[1:]

	// Effective consecutive-failure limit.
	// --max-failure N (>0) wins over --allow-failure when both are set.
	// --allow-failure alone ≡ max-failure 1.
	effectiveMaxFailure := maxFailure
	if maxFailure > 0 {
		effectiveMaxFailure = maxFailure
	} else if allowFailure {
		effectiveMaxFailure = 1
	} else {
		effectiveMaxFailure = 0
	}

	return runLoop(interval, maxRuns, effectiveMaxFailure, command, cmdArgs)
}

func runLoop(interval time.Duration, maxRuns int, effectiveMaxFailure int, command string, cmdArgs []string) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	runs := 0
	consecFail := 0
	var lastExitCode int
	var lastFailed bool

	for {
		// Check for interrupt before starting next run
		select {
		case sig := <-sigChan:
			return &interruptError{Signal: sig}
		default:
		}

		runs++
		exitCode, runErr, interrupted := runOnce(command, cmdArgs, sigChan)
		if interrupted != nil {
			return interrupted
		}

		if runErr != nil {
			fmt.Fprintf(os.Stderr, "for-every: command failed (exit %d): %v\n", exitCode, runErr)
			consecFail++
			lastFailed = true
			lastExitCode = exitCode
			if exitCode == 0 {
				lastExitCode = 1
			}

			if effectiveMaxFailure > 0 && consecFail >= effectiveMaxFailure {
				return errs.NewSilenceExitCode(lastExitCode)
			}
		} else {
			consecFail = 0
			lastFailed = false
			lastExitCode = 0
		}

		if maxRuns > 0 && runs >= maxRuns {
			if lastFailed {
				return errs.NewSilenceExitCode(lastExitCode)
			}
			return nil
		}

		// Sleep between iterations, interruptible
		timer := time.NewTimer(interval)
		select {
		case <-timer.C:
			// continue
		case sig := <-sigChan:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return &interruptError{Signal: sig}
		}
	}
}

func runOnce(command string, args []string, sigChan <-chan os.Signal) (exitCode int, runErr error, interrupted error) {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return 1, fmt.Errorf("failed to start command %q: %v", command, err), nil
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err == nil {
			return 0, nil, nil
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), err, nil
		}
		return 1, err, nil

	case sig := <-sigChan:
		if cmd.Process != nil {
			_ = cmd.Process.Signal(sig)
		}
		// Wait briefly for child to exit after signal
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
				<-done
			}
		}
		return 0, nil, &interruptError{Signal: sig}
	}
}

type interruptError struct {
	Signal os.Signal
}

func (e *interruptError) Error() string {
	return fmt.Sprintf("interrupted by signal %v", e.Signal)
}

func (e *interruptError) SilenceExitCode() int {
	return 130
}
