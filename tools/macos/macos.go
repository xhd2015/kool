package macos

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xhd2015/kool/pkgs/errs"
	"github.com/xhd2015/kool/tools/macos/space"
	lessflags "github.com/xhd2015/less-flags"
)

const help = `macos — macOS-only helpers

Usage:
  kool macos <command> [args...]

Commands:
  space    manage Mission Control Desktops (Spaces)

Options:
  -h, --help    show this help

Run kool macos <command> --help for command-specific options.
`

// Handle runs the kool macos subcommand.
func Handle(args []string) error {
	return run(args, os.Stdout, os.Stderr)
}

// RunForTest runs the handler in-process for tests.
func RunForTest(args []string, stdout, stderr io.Writer) int {
	if err := run(args, stdout, stderr); err != nil {
		if se, ok := errs.IsSilenceExitCode(err); ok {
			return se.SilenceExitCode()
		}
		fmt.Fprint(stderr, err.Error())
		if !strings.HasSuffix(err.Error(), "\n") {
			fmt.Fprintln(stderr)
		}
		return 1
	}
	return 0
}

func run(args []string, stdout, stderr io.Writer) error {
	remain, err := lessflags.
		HelpFunc("-h,--help", func() {}).
		HelpNoExit().
		StopOnFirstArg().
		Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			fmt.Fprint(stdout, strings.TrimSpace(help))
			fmt.Fprintln(stdout)
			return nil
		}
		fmt.Fprint(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}

	if len(remain) == 0 {
		fmt.Fprint(stdout, strings.TrimSpace(help))
		fmt.Fprintln(stdout)
		return nil
	}

	cmd := remain[0]
	sub := remain[1:]
	switch cmd {
	case "space":
		return space.Handle(sub, stdout, stderr)
	case "help", "-h", "--help":
		fmt.Fprint(stdout, strings.TrimSpace(help))
		fmt.Fprintln(stdout)
		return nil
	default:
		fmt.Fprintf(stderr, "unrecognized command: %s\n\n%s\n", cmd, strings.TrimSpace(help))
		return errs.NewSilenceExitCode(1)
	}
}
