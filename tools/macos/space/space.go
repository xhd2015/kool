package space

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	lib "github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/space"
	"github.com/xhd2015/kool/pkgs/errs"
)

const help = `macos space — manage Mission Control Desktops (Spaces)

Usage:
  kool macos space create [--run <cmd> [args...]]
  kool macos space switch <N> [--run <cmd> [args...]]
  kool macos space list

Rule:
  Everything before --run is for the space command.
  Everything after --run is the follow-up command (not parsed as kool flags).

  create              create one Desktop (stay on current Space)
  create --run ...    create, switch to the new Desktop, then run the command
  switch <N>          switch to Desktop N
  switch <N> --run …  switch to Desktop N, then run the command
  list                list Desktop numbers

Requires macOS and Accessibility for the process running kool/osascript.
Mission Control may flash briefly. Not bound to iTerm — pass any follow-up.

Options:
  -h, --help    show this help

Examples:
  kool macos space create
  kool macos space switch 12
  kool macos space list
  kool macos space create --run kool iterm2 -n .
  kool macos space switch 12 --run open -a Safari
`

const createHelp = `macos space create — create one Mission Control Desktop

Usage:
  kool macos space create [--run <cmd> [args...]]

Without --run, creates a Desktop and leaves focus unchanged.
With --run, creates a Desktop, switches to it, settles, then runs the command.

Everything after --run is the follow-up command (not parsed as kool flags).

Options:
  -h, --help    show this help

Examples:
  kool macos space create
  kool macos space create --run kool iterm2 -n .
`

const switchHelp = `macos space switch — switch to a Mission Control Desktop

Usage:
  kool macos space switch <N> [--run <cmd> [args...]]

<N> is the 1-based Desktop number (Desktop 1, Desktop 2, …).

With --run, switches to Desktop N, settles, then runs the command.
Everything after --run is the follow-up command (not parsed as kool flags).

Options:
  -h, --help    show this help

Examples:
  kool macos space switch 12
  kool macos space switch 12 --run kool iterm2 -n .
  kool macos space switch 3 --run open -a Safari
`

const listHelp = `macos space list — list Mission Control Desktops

Usage:
  kool macos space list

Prints one Desktop number per line (1-based), then count=N.

Options:
  -h, --help    show this help

Examples:
  kool macos space list
`

// Handle runs kool macos space (args after "space").
func Handle(args []string, stdout, stderr io.Writer) error {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	return run(args, stdout, stderr)
}

// RunForTest runs the space handler in-process for tests.
func RunForTest(args []string, stdout, stderr io.Writer) int {
	if err := run(args, stdout, stderr); err != nil {
		if se, ok := errs.IsSilenceExitCode(err); ok {
			return se.SilenceExitCode()
		}
		msg := err.Error()
		if msg != "" {
			fmt.Fprint(stderr, msg)
			if !strings.HasSuffix(msg, "\n") {
				fmt.Fprintln(stderr)
			}
		}
		return 1
	}
	return 0
}

func run(args []string, stdout, stderr io.Writer) error {
	left, right, hasRun := splitRun(args)

	if len(left) == 0 {
		printHelp(stdout, help)
		return nil
	}
	if len(left) == 1 && isHelpToken(left[0]) {
		printHelp(stdout, help)
		return nil
	}

	verb := left[0]
	verbArgs := left[1:]

	switch verb {
	case "create":
		return runCreate(verbArgs, right, hasRun, stdout, stderr)
	case "switch":
		return runSwitch(verbArgs, right, hasRun, stdout, stderr)
	case "list":
		return runList(verbArgs, right, hasRun, stdout, stderr)
	case "help", "-h", "--help":
		printHelp(stdout, help)
		return nil
	default:
		fmt.Fprintf(stderr, "unrecognized command: %s\n\n%s\n", verb, strings.TrimSpace(help))
		return errs.NewSilenceExitCode(1)
	}
}

func runCreate(verbArgs, follow []string, hasRun bool, stdout, stderr io.Writer) error {
	if helpOnly, err := rejectExtraOrHelp(verbArgs, createHelp, stdout, stderr); err != nil {
		return err
	} else if helpOnly {
		return nil
	}
	if hasRun && len(follow) == 0 {
		fmt.Fprintln(stderr, "create: --run requires a command")
		return errs.NewSilenceExitCode(1)
	}

	cfg := libConfig()

	if hasRun {
		n, err := createAndActivate(cfg)
		if err != nil {
			return mapLibErr(stderr, "create", err)
		}
		fmt.Fprintf(stdout, "created=true\ndesktop=%d\nswitched=%d\n", n, n)
		return runFollowUp(follow, stdout, stderr)
	}

	if err := doCreate(cfg); err != nil {
		return mapLibErr(stderr, "create", err)
	}
	h, err := doHighest(cfg)
	if err != nil {
		fmt.Fprintf(stderr, "create: created but could not resolve new desktop: %v\n", err)
		return errs.NewSilenceExitCode(1)
	}
	fmt.Fprintf(stdout, "created=true\ndesktop=%d\n", h)
	return nil
}

func runSwitch(verbArgs, follow []string, hasRun bool, stdout, stderr io.Writer) error {
	if containsHelp(verbArgs) {
		printHelp(stdout, switchHelp)
		return nil
	}
	if len(verbArgs) == 0 {
		fmt.Fprintln(stderr, "switch: missing desktop number")
		fmt.Fprintln(stderr)
		fmt.Fprint(stderr, strings.TrimSpace(switchHelp))
		fmt.Fprintln(stderr)
		return errs.NewSilenceExitCode(1)
	}
	if len(verbArgs) > 1 {
		fmt.Fprintf(stderr, "switch: unexpected arguments: %s\n", strings.Join(verbArgs[1:], " "))
		return errs.NewSilenceExitCode(1)
	}
	n, err := parseDesktopNumber(verbArgs[0])
	if err != nil {
		fmt.Fprintf(stderr, "switch: %v\n", err)
		return errs.NewSilenceExitCode(1)
	}
	if hasRun && len(follow) == 0 {
		fmt.Fprintln(stderr, "switch: --run requires a command")
		return errs.NewSilenceExitCode(1)
	}

	cfg := libConfig()
	if err := doSwitch(n, cfg); err != nil {
		return mapLibErr(stderr, "switch", err)
	}
	fmt.Fprintf(stdout, "switched=%d\n", n)

	if !hasRun {
		return nil
	}
	return runFollowUp(follow, stdout, stderr)
}

func runList(verbArgs, follow []string, hasRun bool, stdout, stderr io.Writer) error {
	if hasRun {
		fmt.Fprintln(stderr, "list: --run is not supported")
		return errs.NewSilenceExitCode(1)
	}
	if helpOnly, err := rejectExtraOrHelp(verbArgs, listHelp, stdout, stderr); err != nil {
		return err
	} else if helpOnly {
		return nil
	}

	desktops, err := doList(libConfig())
	if err != nil {
		return mapLibErr(stderr, "list", err)
	}
	for _, d := range desktops {
		fmt.Fprintln(stdout, d.Number)
	}
	fmt.Fprintf(stdout, "count=%d\n", len(desktops))
	return nil
}

func runFollowUp(follow []string, stdout, stderr io.Writer) error {
	if len(follow) == 0 {
		fmt.Fprintln(stderr, "--run requires a command")
		return errs.NewSilenceExitCode(1)
	}
	// Settle already applied by library Switch / CreateAndActivate when using default cfg.
	// Extra settle only if caller used mock without settle — keep zero extra here.
	name := follow[0]
	args := follow[1:]
	fmt.Fprintf(stdout, "run=%s\n", strings.Join(follow, " "))
	if err := getRunner().Run(name, args); err != nil {
		if se, ok := errs.IsSilenceExitCode(err); ok {
			return se
		}
		fmt.Fprintf(stderr, "follow-up: %v\n", err)
		return errs.NewSilenceExitCode(1)
	}
	return nil
}

func mapLibErr(stderr io.Writer, op string, err error) error {
	if errors.Is(err, lib.ErrUnsupportedPlatform) {
		fmt.Fprintln(stderr, "kool macos space is only supported on macOS")
		return errs.NewSilenceExitCode(1)
	}
	fmt.Fprintf(stderr, "%s: %v\n", op, err)
	return errs.NewSilenceExitCode(1)
}

func printHelp(w io.Writer, text string) {
	fmt.Fprint(w, strings.TrimSpace(text))
	fmt.Fprintln(w)
}

func containsHelp(args []string) bool {
	for _, a := range args {
		if isHelpToken(a) {
			return true
		}
	}
	return false
}

func rejectExtraOrHelp(args []string, levelHelp string, stdout, stderr io.Writer) (helpOnly bool, err error) {
	if len(args) == 0 {
		return false, nil
	}
	if containsHelp(args) {
		printHelp(stdout, levelHelp)
		return true, nil
	}
	fmt.Fprintf(stderr, "unexpected arguments: %s\n", strings.Join(args, " "))
	return false, errs.NewSilenceExitCode(1)
}

// libConfig builds Config for library calls (tests may inject backend path via hooks).
func libConfig() *lib.Config {
	cfg := &lib.Config{}
	if ms := settleMS(); ms < 0 {
		cfg.Settle = -1
	} else if ms == 0 {
		cfg.Settle = -1 // tests: no sleep
	} else {
		cfg.Settle = time.Duration(ms) * time.Millisecond
	}
	if osascript := testOsascript(); osascript != nil {
		cfg.Osascript = osascript
	}
	return cfg
}
