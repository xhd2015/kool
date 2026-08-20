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
iterm2 focus <dir> [--index N]
iterm2 install [--dry-run] [--download-dir DIR] [--download-only] [--via-open]
iterm2 tab-set list|show|run|status|stop ...
iterm2 sessions snapshot|save|restore [options]
iterm2 session <session-id> status [options]
iterm2 contents <session-id> [options]

Open a directory in iTerm2 on macOS, get/set the current session or window title
when running inside iTerm2 (ITERM_SESSION_ID set), manage named tab-set layouts,
install official iTerm2 (no Homebrew), or snapshot / save / restore live sessions.

Open directory:
  dir                              directory to open (required)
  -r, --reuse, --reuse-window      focus session at dir if open; else new window + cd
  -n, --new-window                 always open in a new window (cannot use with -r)
  --send <command>                 shell command to run after cd (repeatable)

Title commands (require ITERM_SESSION_ID):
  set-title [--window] <title>     set session/tab name, or window name with --window
  get-title [--window]             print session/tab name, or window name with --window

Focus command:
  focus <dir> [--index N]          focus one existing exact directory match; never creates a session

Install official iTerm2 (zip from iterm2.com → ~/Applications by default):
  install [flags]                  resolve/download/install latest stable
  install --dry-run                print plan only (no download/install)
  install --download-dir <dir>     zip/cache directory
  install --download-only          download zip only (no Applications install)
  install --via-open               user-driven install (clear quarantine + open / Gatekeeper)

Tab sets (config: ~/.config/iterm2/tab-set or KOOL_ITERM2_TAB_SET_DIR):
  tab-set list                     list configured tab sets
  tab-set show <name>              show tabs/commands for a set
  tab-set run <name> [flags]       run or resync a tab set (--dry-run, -n, --no-new-window)
  tab-set status <name>            report running/idle/missing tabs
  tab-set stop <name>              close marked windows/tabs for a set
  tab-set -h|--help                tab-set usage

Sessions snapshot / save / restore / auto-backup / status:
  sessions snapshot [options]      dump all windows/tabs/sessions (cli|json|md|html)
  sessions save [--dry-run] [--file PATH]
                                   checkpoint critical grok/codex/mark tabs
  sessions restore [--dry-run] [--file PATH] [--same-app]
                                   recreate windows and resume from checkpoint
  sessions auto-backup [--once] [--interval DUR] [--file PATH]
                                   periodically checkpoint critical tabs (default 10m)
  session <id> status [options]    live status for one session (id = iTerm unique ID)
  contents <session-id>            print visible pane text (no focus; home then system app)

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
  kool iterm2 install --dry-run
  kool iterm2 install --download-only --download-dir ~/Downloads
  kool iterm2 install --via-open
  kool iterm2 sessions snapshot
  kool iterm2 sessions snapshot --json -o iterm.json
  kool iterm2 sessions save --dry-run
  kool iterm2 sessions save
  kool iterm2 sessions restore
  kool iterm2 sessions auto-backup --once
  kool iterm2 session D922B298 status
  kool iterm2 contents B95E6BAC-3104-43D2-ABAE-86FC02A669A2
  kool iterm2 contents B95E6BAC-3104-43D2-ABAE-86FC02A669A2 --json
`

// SetGOOSForTest overrides platform detection for handler tests.
func SetGOOSForTest(goos string) {
	lib.SetGOOSForTest(goos)
}

// TestRun injects handler-test overrides. Do not mutate process env or cwd.
type TestRun struct {
	TabSetDir string
	GOOS      string
	Installed func() bool
	Osascript func(script string) error
	// Contents overrides lib.Contents (tests).
	Contents func(sessionID string, cfg *lib.ContentsConfig) (lib.ContentsResult, error)
}

// Handle runs the kool iterm2 subcommand.
func Handle(args []string) error {
	return run(args, os.Stdout, os.Stderr, TestRun{})
}

// RunForTest runs the handler in-process for doctest handler phase.
// workingDir is unused; callers pass absolute paths in args.
func RunForTest(args []string, stdout, stderr io.Writer, workingDir string) int {
	_ = workingDir
	return RunForTestEnv(args, stdout, stderr, TestRun{})
}

// RunForTestEnv is RunForTest with explicit hooks (no process Setenv/Chdir).
func RunForTestEnv(args []string, stdout, stderr io.Writer, env TestRun) int {
	if env.GOOS != "" {
		lib.SetGOOSForTest(env.GOOS)
	}
	if err := run(args, stdout, stderr, env); err != nil {
		return 1
	}
	return 0
}

func run(args []string, stdout, stderr io.Writer, env TestRun) error {
	// Reserved first-arg routing for title / tab-set / sessions (before open-dir).
	if len(args) > 0 {
		switch args[0] {
		case "set-title":
			return runSetTitle(args[1:], stdout, stderr)
		case "get-title":
			return runGetTitle(args[1:], stdout, stderr)
		case "focus":
			if runFocus(args[1:], stdout, stderr, liveFocusBoundary{}) != 0 {
				return errs.NewSilenceExitCode(1)
			}
			return nil
		case "install":
			return runInstall(args[1:], stdout, stderr)
		case "tab-set":
			return runTabSet(args[1:], stdout, stderr, env.TabSetDir)
		case "sessions":
			return runSessions(args[1:], stdout, stderr)
		case "session":
			return runSession(args[1:], stdout, stderr)
		case "contents":
			return runContents(args[1:], stdout, stderr, env)
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
			fmt.Fprint(stdout, strings.TrimSpace(help)+"\n")
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
	cfg := &lib.Config{
		FollowUpCommands: sends,
		SafeInputIgnore:  true,
	}
	if env.Installed != nil {
		cfg.Installed = env.Installed
	}
	if env.Osascript != nil {
		cfg.Osascript = env.Osascript
	}
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
			fmt.Fprint(stdout, strings.TrimSpace(help)+"\n")
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
			fmt.Fprint(stdout, strings.TrimSpace(help)+"\n")
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
