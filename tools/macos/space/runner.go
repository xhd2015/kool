package space

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/xhd2015/kool/pkgs/errs"
)

// Runner executes a --run follow-up command.
type Runner interface {
	Run(name string, args []string) error
}

type execRunner struct{}

func (execRunner) Run(name string, args []string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err == nil {
		return nil
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return errs.NewSilenceExitCode(ee.ExitCode())
	}
	return fmt.Errorf("run %s: %w", name, err)
}

// RecordingRunner records follow-up invocations (tests).
type RecordingRunner struct {
	Calls [][]string
	Err   error
}

func (r *RecordingRunner) Run(name string, args []string) error {
	call := append([]string{name}, args...)
	r.Calls = append(r.Calls, call)
	return r.Err
}
