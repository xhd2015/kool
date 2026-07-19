package iterm2

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	lib "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	"github.com/xhd2015/kool/pkgs/errs"
	lessflags "github.com/xhd2015/less-flags"
)

const help = `iterm2 <dir> [-r] [-n] [--send <command>]...
iterm2 set-title [--window] <title>
iterm2 get-title [--window]
iterm2 tab-set list|show|run|status|stop ...

Open a directory in iTerm2 on macOS, get/set the current session or window title
when running inside iTerm2 (ITERM_SESSION_ID set), or manage named tab-set layouts.

Open directory:
  dir                              directory to open (required)
  -r, --reuse, --reuse-window      focus session at dir if open; else new window + cd
  -n, --new-window                 always open in a new window (cannot use with -r)
  --send <command>                 shell command to run after cd (repeatable)

Title commands (require ITERM_SESSION_ID):
  set-title [--window] <title>     set session/tab name, or window name with --window
  get-title [--window]             print session/tab name, or window name with --window

Tab sets (config: ~/.config/iterm2/tab-set or KOOL_ITERM2_TAB_SET_DIR):
  tab-set list                     list configured tab sets
  tab-set show <name>              show tabs/commands for a set
  tab-set run <name> [flags]       run or resync a tab set (--dry-run, -n, --no-new-window)
  tab-set status <name>            report running/idle/missing tabs
  tab-set stop <name>              close marked windows/tabs for a set
  tab-set -h|--help                tab-set usage

Options:
  -h, --help                       show this help message

Examples:
  kool iterm2 .
  kool iterm2 /path/to/project -r
  kool iterm2 /path/to/project -n
  kool iterm2 /path/to/project --send grok
  kool iterm2 . --send grok --send codex
  kool iterm2 set-title my-title
  kool iterm2 set-title --window "Project Window"
  kool iterm2 get-title
  kool iterm2 get-title --window
  kool iterm2 tab-set list
  kool iterm2 tab-set run bots --dry-run
`

// SetGOOSForTest overrides platform detection for handler tests.
func SetGOOSForTest(goos string) {
	lib.SetGOOSForTest(goos)
}

// Handle runs the kool iterm2 subcommand.
func Handle(args []string) error {
	return run(args, os.Stdout, os.Stderr)
}

// RunForTest runs the handler in-process for doctest handler phase.
func RunForTest(args []string, stdout, stderr io.Writer, workingDir string) int {
	prev, _ := os.Getwd()
	if workingDir != "" {
		if err := os.Chdir(workingDir); err != nil {
			fmt.Fprintf(stderr, "chdir: %v\n", err)
			return 1
		}
		defer func() { _ = os.Chdir(prev) }()
	}
	if err := run(args, stdout, stderr); err != nil {
		return 1
	}
	return 0
}

func run(args []string, stdout, stderr io.Writer) error {
	// Reserved first-arg routing for title / tab-set (before open-dir).
	if len(args) > 0 {
		switch args[0] {
		case "set-title":
			return runSetTitle(args[1:], stdout, stderr)
		case "get-title":
			return runGetTitle(args[1:], stdout, stderr)
		case "tab-set":
			return runTabSet(args[1:], stdout, stderr)
		}
	}

	var sends []string
	var reuse bool
	var newWindow bool
	remain, err := lessflags.StringSlice("--send", &sends).
		Bool("-r,--reuse,--reuse-window", &reuse).
		Bool("-n,--new-window", &newWindow).
		HelpFunc("-h,--help", func() {}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			fmt.Fprint(stdout, strings.TrimSpace(help))
			return nil
		}
		fmt.Fprint(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}

	if newWindow && reuse {
		fmt.Fprintf(stderr, "cannot specify both -n/--new-window and -r/--reuse/--reuse-window")
		return errs.NewSilenceExitCode(1)
	}

	if len(remain) == 0 {
		fmt.Fprintf(stderr, "missing directory argument\n\n%s", strings.TrimSpace(help))
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) > 1 {
		fmt.Fprintf(stderr, "unrecognized arguments: %s", strings.Join(remain[1:], " "))
		return errs.NewSilenceExitCode(1)
	}

	dir := remain[0]
	cfg := &lib.Config{FollowUpCommands: sends}
	if newWindow {
		cfg.Mode = lib.ModeForceNew
	} else if reuse {
		cfg.Mode = lib.ModeReuseCurrent
	}
	if err := lib.OpenConfig(dir, cfg); err != nil {
		if errors.Is(err, lib.ErrUnsupportedPlatform) {
			fmt.Fprint(stderr, "Open i2Term2 is only supported on macOS.")
			return errs.NewSilenceExitCode(1)
		}
		fmt.Fprint(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	return nil
}

func runSetTitle(args []string, stdout, stderr io.Writer) error {
	var window bool
	remain, err := lessflags.Bool("--window", &window).
		HelpFunc("-h,--help", func() {}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			fmt.Fprint(stdout, strings.TrimSpace(help))
			return nil
		}
		fmt.Fprint(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}

	if len(remain) == 0 {
		fmt.Fprint(stderr, "set-title: missing title argument\n")
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) > 1 {
		fmt.Fprintf(stderr, "set-title: unexpected arguments: %s\n", strings.Join(remain[1:], " "))
		return errs.NewSilenceExitCode(1)
	}

	title := remain[0]
	if title == "" {
		fmt.Fprint(stderr, "set-title: title must not be empty\n")
		return errs.NewSilenceExitCode(1)
	}

	target := lib.TitleTargetSession
	if window {
		target = lib.TitleTargetWindow
	}

	old, newTitle, err := lib.SetTitle(title, target)
	if err != nil {
		if errors.Is(err, lib.ErrNotInSession) {
			fmt.Fprint(stderr, "warning: nothing to set; needs to be run inside iTerm2\n")
			return errs.NewSilenceExitCode(1)
		}
		if errors.Is(err, lib.ErrEmptyTitle) {
			fmt.Fprint(stderr, "set-title: title must not be empty\n")
			return errs.NewSilenceExitCode(1)
		}
		fmt.Fprint(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}

	fmt.Fprintf(stdout, "title changed: %s -> %s\n", old, newTitle)
	return nil
}

func runGetTitle(args []string, stdout, stderr io.Writer) error {
	var window bool
	remain, err := lessflags.Bool("--window", &window).
		HelpFunc("-h,--help", func() {}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			fmt.Fprint(stdout, strings.TrimSpace(help))
			return nil
		}
		fmt.Fprint(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}

	if len(remain) > 0 {
		fmt.Fprintf(stderr, "get-title: unexpected arguments: %s\n", strings.Join(remain, " "))
		return errs.NewSilenceExitCode(1)
	}

	target := lib.TitleTargetSession
	if window {
		target = lib.TitleTargetWindow
	}

	title, err := lib.GetTitle(target)
	if err != nil {
		if errors.Is(err, lib.ErrNotInSession) {
			fmt.Fprint(stderr, "warning: nothing to get; needs to be run inside iTerm2\n")
			return errs.NewSilenceExitCode(1)
		}
		fmt.Fprint(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}

	fmt.Fprintln(stdout, title)
	return nil
}
