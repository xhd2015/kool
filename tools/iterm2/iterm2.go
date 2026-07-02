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

Open a directory in iTerm2 on macOS.

Arguments:
dir                              directory to open (required)

Options:
-r, --reuse, --reuse-window      focus session at dir if open; else new window + cd
-n, --new-window                 always open in a new window (cannot use with -r)
--send <command>                 shell command to run after cd (repeatable)
-h, --help                       show this help message

Examples:
kool iterm2 .
kool iterm2 /path/to/project -r
kool iterm2 /path/to/project -n
kool iterm2 /path/to/project --send grok
kool iterm2 . --send grok --send codex
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