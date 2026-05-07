package go_tools

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestHandleRefactorHelp(t *testing.T) {
	for _, args := range [][]string{
		{"--help"},
		{"-h"},
	} {
		output := runHandleRefactor(t, args...)
		if !strings.Contains(output, "Usage: kool go refactor <command> [OPTIONS]") {
			t.Fatalf("HandleRefactor(%v) output missing usage:\n%s", args, output)
		}
		if !strings.Contains(output, "move <src> <dst>") {
			t.Fatalf("HandleRefactor(%v) output missing move command:\n%s", args, output)
		}
	}
}

func TestHandleRefactorHelpCommandIsNotSpecial(t *testing.T) {
	err := HandleRefactor([]string{"help"})
	if err == nil {
		t.Fatal("HandleRefactor(help) returned nil")
	}
	if err.Error() != "unknown command: help" {
		t.Fatalf("HandleRefactor(help) error = %q", err)
	}
}

func TestHandleRefactorMoveHelp(t *testing.T) {
	output := runHandleRefactor(t, "move", "--help")
	if !strings.Contains(output, "Usage: kool go refactor move <src> <dst>") {
		t.Fatalf("HandleRefactor(move --help) output missing usage:\n%s", output)
	}
	if strings.Contains(output, "Commands:") {
		t.Fatalf("HandleRefactor(move --help) printed parent help instead of move help:\n%s", output)
	}
}

func TestHandleRefactorHelperProcess(t *testing.T) {
	if os.Getenv("KOOL_TEST_HANDLE_REFACTOR") != "1" {
		return
	}
	args := os.Args
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}
	if err := HandleRefactor(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func runHandleRefactor(t *testing.T, args ...string) string {
	t.Helper()

	cmdArgs := []string{"-test.run=^TestHandleRefactorHelperProcess$", "--"}
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command(os.Args[0], cmdArgs...)
	cmd.Env = append(os.Environ(), "KOOL_TEST_HANDLE_REFACTOR=1")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("HandleRefactor(%v) failed: %v\n%s", args, err, output)
	}
	return string(output)
}
