package iterm2

import (
	"errors"
	"fmt"
	"io"
	"strings"

	lib "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	"github.com/xhd2015/kool/pkgs/errs"
	lessflags "github.com/xhd2015/less-flags"
)

const windowHelp = `iterm2 window — status for the iTerm2 window that parents this process

Usage:
  kool iterm2 window status
  kool iterm2 window -h|--help

Commands:
  status    list all tabs in the current window; mark * on this process's tab

Resolves the parent pane via ITERM_SESSION_ID, else a controlling/ancestor TTY
that appears in the live iTerm session list.
`

const tabHelp = `iterm2 tab — status for the iTerm2 tab that parents this process

Usage:
  kool iterm2 tab status
  kool iterm2 tab -h|--help

Commands:
  status    show the current tab (window id/name, tab index/name, session, tty)

Resolves the parent pane via ITERM_SESSION_ID, else a controlling/ancestor TTY
that appears in the live iTerm session list.
`

func runWindow(args []string, stdout, stderr io.Writer, env TestRun) error {
	if len(args) == 0 {
		fmt.Fprint(stdout, strings.TrimSpace(windowHelp)+"\n")
		return nil
	}
	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Fprint(stdout, strings.TrimSpace(windowHelp)+"\n")
		return nil
	}
	switch args[0] {
	case "status":
		return runWindowStatus(args[1:], stdout, stderr, env)
	default:
		fmt.Fprintf(stderr, "Error: window: unknown command %q (expected status)\n\n%s\n", args[0], strings.TrimSpace(windowHelp))
		return errs.NewSilenceExitCode(1)
	}
}

func runTab(args []string, stdout, stderr io.Writer, env TestRun) error {
	if len(args) == 0 {
		fmt.Fprint(stdout, strings.TrimSpace(tabHelp)+"\n")
		return nil
	}
	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Fprint(stdout, strings.TrimSpace(tabHelp)+"\n")
		return nil
	}
	switch args[0] {
	case "status":
		return runTabStatus(args[1:], stdout, stderr, env)
	default:
		fmt.Fprintf(stderr, "Error: tab: unknown command %q (expected status)\n\n%s\n", args[0], strings.TrimSpace(tabHelp))
		return errs.NewSilenceExitCode(1)
	}
}

func runWindowStatus(args []string, stdout, stderr io.Writer, env TestRun) error {
	remain, err := lessflags.HelpFunc("-h,--help", func() {}).HelpNoExit().Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			fmt.Fprint(stdout, strings.TrimSpace(windowHelp)+"\n")
			return nil
		}
		fmt.Fprint(stderr, "Error: "+err.Error()+"\n")
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) > 0 {
		fmt.Fprintf(stderr, "Error: window status: unexpected arguments: %s\n", strings.Join(remain, " "))
		return errs.NewSilenceExitCode(1)
	}

	st, err := lib.CurrentWindowStatusWith(env.currentStatusConfig())
	if err != nil {
		return writeCurrentStatusErr(stderr, err)
	}
	fmt.Fprint(stdout, lib.FormatWindowStatus(st))
	return nil
}

func runTabStatus(args []string, stdout, stderr io.Writer, env TestRun) error {
	remain, err := lessflags.HelpFunc("-h,--help", func() {}).HelpNoExit().Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			fmt.Fprint(stdout, strings.TrimSpace(tabHelp)+"\n")
			return nil
		}
		fmt.Fprint(stderr, "Error: "+err.Error()+"\n")
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) > 0 {
		fmt.Fprintf(stderr, "Error: tab status: unexpected arguments: %s\n", strings.Join(remain, " "))
		return errs.NewSilenceExitCode(1)
	}

	st, err := lib.CurrentTabStatusWith(env.currentStatusConfig())
	if err != nil {
		return writeCurrentStatusErr(stderr, err)
	}
	fmt.Fprint(stdout, lib.FormatTabStatus(st))
	return nil
}

func writeCurrentStatusErr(stderr io.Writer, err error) error {
	if errors.Is(err, lib.ErrNotInSession) {
		fmt.Fprint(stderr, "Error: iterm2: not inside an iTerm2 session (no ITERM_SESSION_ID and no matching TTY)\n")
		return errs.NewSilenceExitCode(1)
	}
	fmt.Fprint(stderr, "Error: "+err.Error()+"\n")
	return errs.NewSilenceExitCode(1)
}
