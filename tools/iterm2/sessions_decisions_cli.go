package iterm2

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xhd2015/kool/pkgs/errs"
	lessflags "github.com/xhd2015/less-flags"
)

const sessionsDecisionsHelp = `iterm2 sessions decisions — list or remove saved restore decisions

Usage:
  kool iterm2 sessions decisions list [--file PATH]
  kool iterm2 sessions decisions rm <index|key> [--file PATH]
  kool iterm2 sessions decisions -h|--help

Saved restore decisions remember your allow/deny choice for an exact command
(cwd + argv). They are applied automatically on later restores, with a notice
printed. list shows them; rm forgets one by list index or by its exact key.

  --file PATH   decisions file (default: ~/.config/iterm2/restore-decisions.json)
  -h, --help

Examples:
  kool iterm2 sessions decisions list
  kool iterm2 sessions decisions rm 2
`

// runSessionsDecisions implements `sessions decisions list|rm`.
func runSessionsDecisions(args []string, stdout, stderr io.Writer) error {
	var fileFlag string
	remain, err := lessflags.String("-f,--file", &fileFlag).
		HelpFunc("-h,--help", func() {}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			fmt.Fprint(stdout, strings.TrimSpace(sessionsDecisionsHelp)+"\n")
			return nil
		}
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) == 0 {
		fmt.Fprint(stdout, strings.TrimSpace(sessionsDecisionsHelp)+"\n")
		return nil
	}
	path := fileFlag
	if path == "" {
		path = effectiveDecisionsPath()
	}
	store, warns := ReadDecisionsDocument(path)
	for _, w := range warns {
		WriteWarning(stderr, w)
	}
	switch remain[0] {
	case "list":
		listDecisions(stdout, store)
		return nil
	case "rm":
		if len(remain) < 2 {
			WriteError(stderr, "sessions decisions rm: missing <index|key>")
			return errs.NewSilenceExitCode(1)
		}
		return rmDecision(path, remain[1], stdout, stderr, store)
	default:
		fmt.Fprintf(stderr, "Error: sessions decisions: unknown subcommand %q\n\n%s\n", remain[0], strings.TrimSpace(sessionsDecisionsHelp))
		return errs.NewSilenceExitCode(1)
	}
}

func rmDecision(path, selector string, stdout, stderr io.Writer, store *DecisionsDocument) error {
	key := selector
	if idx, aerr := strconv.Atoi(selector); aerr == nil && idx >= 1 && idx <= len(store.Decisions) {
		key = store.Decisions[idx-1].Key
	}
	if !removeDecision(store, key) {
		WriteError(stderr, fmt.Sprintf("sessions decisions rm: no decision matches %q", selector))
		return errs.NewSilenceExitCode(1)
	}
	if err := WriteDecisionsDocument(path, store); err != nil {
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	fmt.Fprintf(stdout, "removed 1 decision\n")
	return nil
}
