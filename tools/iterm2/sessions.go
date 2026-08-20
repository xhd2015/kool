package iterm2

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xhd2015/kool/pkgs/errs"
	lessflags "github.com/xhd2015/less-flags"
)

const sessionsHelp = `iterm2 sessions — inspect live iTerm2 windows, tabs, and sessions

Usage:
  kool iterm2 sessions snapshot [options]
  kool iterm2 sessions save [--dry-run] [--file PATH] [--color|--no-color] [--ignore-macos-space] [--spaces LIST]
  kool iterm2 sessions restore [--dry-run] [--file PATH] [--color|--no-color] [--ignore-macos-space] [--same-app]
  kool iterm2 sessions auto-backup [--once] [--interval DUR] [--file PATH] [--dry-run]
  kool iterm2 sessions -h|--help

Commands:
  snapshot                 capture all windows/tabs/sessions for human review
  save                     checkpoint critical grok/codex/mark tabs for restore
  restore                  recreate windows and resume from the last save file
  auto-backup              periodically checkpoint critical tabs (default every 10m)

Snapshot options:
  --json                   emit JSON
  --markdown               emit Markdown
  --html                   emit HTML
  -o, --output FILE        write to FILE (format inferred from .json/.md/.html
                           when no format flag is set)
  --no-stream              buffer full CLI (default streams window blocks)
  --no-color               disable ANSI colors on CLI output
  --no-enrich              skip agent session resolve (no grok/codex session id)
  --no-tree                keep agent session id but omit process tree lines
  -h, --help               show this help

Save / restore / auto-backup:
  Manual checkpoint default: ~/.config/iterm2/sessions-save.json
  Auto-backup default: ~/.config/iterm2/sessions-auto.json (always overwrite)
  save keeps panes with resolved grok/codex session ids, or a live mark process.
  restore creates windows/tabs, then: cd <cwd> and grok --resume / codex resume / mark …
  auto-backup loops (default every 10m) with --once for a single cycle; soft-fails capture.
  If the checkpoint exists and is not restored yet, save prompts [Y/n] on a TTY
  (non-TTY stdin errors). restore errors if restored_at is already set.
  --color / --no-color force ANSI on or off for save/restore/auto-backup plan output.

Default format is colored CLI (when stdout is a TTY). CLI streams each window
block as it is collected; use --no-stream to buffer. JSON/Markdown/HTML always
buffer one document. On busy panes, agent-pro procresolve attaches grok/codex
session ids and a Unicode process tree unless --no-enrich / --no-tree.
Session id is the iTerm2 session unique ID (UUID); see also:
  kool iterm2 session <session-id> status

Examples:
  kool iterm2 sessions snapshot
  kool iterm2 sessions snapshot --no-stream
  kool iterm2 sessions snapshot --json
  kool iterm2 sessions snapshot -o ~/Desktop/iterm.json
  kool iterm2 sessions snapshot --markdown -o review.md
  kool iterm2 sessions snapshot --no-enrich
  kool iterm2 sessions snapshot --no-tree
  kool iterm2 sessions save --dry-run
  kool iterm2 sessions save
  kool iterm2 sessions restore --dry-run
  kool iterm2 sessions restore
  kool iterm2 sessions auto-backup --once
  kool iterm2 sessions auto-backup --interval 5m
`

const sessionsSaveHelp = `iterm2 sessions save — checkpoint critical grok/codex/mark tabs

Usage: kool iterm2 sessions save [--dry-run] [--file PATH] [--color|--no-color] [--ignore-macos-space] [--spaces LIST]

Save busy panes that have a resolved grok or codex session_id, or a live mark
process, into a checkpoint JSON (default: ~/.config/iterm2/sessions-save.json).

  --dry-run              print plan only; do not write or prompt
  --file PATH            checkpoint path (default: ~/.config/iterm2/sessions-save.json)
  --color                force ANSI colors on (wins over NO_COLOR / non-TTY)
  --no-color             force ANSI colors off
  --ignore-macos-space   omit space / iterm_window_id; do not resolve Spaces
  --spaces LIST          only save windows on these 0-based Space indexes
                         (comma-separated, e.g. 0,2); cannot combine with
                         --ignore-macos-space; recorded as filter.spaces
  -h, --help

Overwrite: if the file exists and has not been restored yet, prompts [Y/n]
on a TTY; non-interactive stdin errors out. Already-restored files are
overwritten without a prompt. Zero critical sessions: exit 0, no write.
Dry-run streams each critical window block as it is captured, then a footer.
By default each window records macOS Space (space + iterm_window_id).
With --spaces, windows outside the list are omitted; a skip warning is
printed when any critical window was dropped.

Multi-app: when dual iTerm installs are running (~/Applications/iTerm.app and
/Applications/iTerm.app), save merges windows from both and records per-window
canonical app. Restore by default prefers ~/Applications/iTerm.app when both
installs exist; use restore --same-app to recreate each window in its recorded app.

Examples:
  kool iterm2 sessions save --dry-run
  kool iterm2 sessions save
  kool iterm2 sessions save --file ~/Desktop/pre-reboot.json
  kool iterm2 sessions save --ignore-macos-space
  kool iterm2 sessions save --spaces 0,2
`

const sessionsRestoreHelp = `iterm2 sessions restore — recreate windows and resume from checkpoint

Usage: kool iterm2 sessions restore [--dry-run] [--file PATH] [--color|--no-color] [--ignore-macos-space] [--same-app]

Read the checkpoint, create one window per saved window, one tab per entry,
then send: cd <cwd> and grok --resume / codex resume / mark <message>.

  --dry-run              print plan only; do not create tabs or mark restored
  --file PATH            checkpoint path (default: ~/.config/iterm2/sessions-save.json)
  --color                force ANSI colors on (wins over NO_COLOR / non-TTY)
  --no-color             force ANSI colors off
  --ignore-macos-space   ignore recorded space; create on current Desktop
  --same-app             recreate each window in its recorded app (canonical
                         home/system path). Default prefers
                         ~/Applications/iTerm.app when multiple installs exist;
                         else the only install on disk; else bare iTerm2
  -h, --help

If restored_at is set, the file is already consumed and restore errors.
On full success, restored_at is written so the checkpoint cannot be applied twice.
By default each window is placed on its recorded macOS Space (Switch/Create).
Create target: by default one global install (prefer home when both exist);
--same-app uses each window's recorded app.

Examples:
  kool iterm2 sessions restore --dry-run
  kool iterm2 sessions restore
  kool iterm2 sessions restore --file ~/Desktop/pre-reboot.json
  kool iterm2 sessions restore --ignore-macos-space
  kool iterm2 sessions restore --same-app
`

const sessionHelp = `iterm2 session — inspect a single iTerm2 session

Usage:
  kool iterm2 session <session-id> status [options]
  kool iterm2 session -h|--help

Session id:
  Primary: iTerm2 session unique ID (UUID from snapshot "id" field).
  Also accepted when unique among live sessions:
    UUID prefix (≥8 chars), tty (ttys003 or /dev/ttys003), or pid.

Commands:
  status                   re-query live status for the session

Status options:
  --json                   emit JSON
  --markdown               emit Markdown
  --html                   emit HTML
  -o, --output FILE        write to FILE (suffix may infer format)
  --no-color               disable ANSI colors on CLI output
  --no-enrich              skip agent session resolve (no grok/codex session id)
  --no-tree                keep agent session id but omit process tree lines
  -h, --help               show this help

Examples:
  kool iterm2 session D922B298-25FB-41FA-BAF8-7AC7A1D56758 status
  kool iterm2 session D922B298 status --json
  kool iterm2 session ttys003 status
`

func runSessions(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stdout, strings.TrimSpace(sessionsHelp)+"\n")
		return nil
	}
	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Fprint(stdout, strings.TrimSpace(sessionsHelp)+"\n")
		return nil
	}
	switch args[0] {
	case "snapshot":
		return runSessionsSnapshot(args[1:], stdout, stderr)
	case "save":
		return runSessionsSave(args[1:], stdout, stderr)
	case "restore":
		return runSessionsRestore(args[1:], stdout, stderr)
	case "auto-backup":
		return runSessionsAutoBackup(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "Error: sessions: unknown subcommand %q\n\n%s\n", args[0], strings.TrimSpace(sessionsHelp))
		return errs.NewSilenceExitCode(1)
	}
}

func runSessionsSnapshot(args []string, stdout, stderr io.Writer) error {
	var asJSON, asMD, asHTML, noColor, noStream, noEnrich, noTree bool
	var outPath string
	remain, err := lessflags.Bool("--json", &asJSON).
		Bool("--markdown", &asMD).
		Bool("--html", &asHTML).
		Bool("--no-color", &noColor).
		Bool("--no-stream", &noStream).
		Bool("--no-enrich", &noEnrich).
		Bool("--no-tree", &noTree).
		String("-o,--output", &outPath).
		HelpFunc("-h,--help", func() {}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			fmt.Fprint(stdout, strings.TrimSpace(sessionsHelp)+"\n")
			return nil
		}
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) > 0 {
		fmt.Fprintf(stderr, "Error: sessions snapshot: unexpected arguments: %s\n", strings.Join(remain, " "))
		return errs.NewSilenceExitCode(1)
	}

	format, err := ResolveFormat(FormatFlags{JSON: asJSON, Markdown: asMD, HTML: asHTML}, outPath)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}

	capOpts := CaptureOpts{NoEnrich: noEnrich}

	// Progressive CLI: stream each window block after enrich; footer last.
	// JSON/MD/HTML, -o, and --no-stream stay fully buffered.
	streamCLI := format == FormatCLI && outPath == "" && !noStream
	if streamCLI {
		return runSessionsSnapshotStream(stdout, stderr, noColor, noTree, capOpts)
	}

	snap, warnings, err := CaptureSnapshotWith(capOpts)
	if err != nil {
		WriteError(stderr, strings.TrimPrefix(err.Error(), "Error: "))
		return errs.NewSilenceExitCode(1)
	}
	for _, w := range warnings {
		WriteWarning(stderr, strings.TrimPrefix(w, "warning: "))
	}

	opt := RenderOptions{Format: format, NoColor: noColor, NoTree: noTree}
	if outPath != "" {
		f, err := os.Create(outPath)
		if err != nil {
			WriteError(stderr, err.Error())
			return errs.NewSilenceExitCode(1)
		}
		defer f.Close()
		// never color files
		opt.NoColor = true
		if err := RenderSnapshot(f, snap, opt); err != nil {
			WriteError(stderr, err.Error())
			return errs.NewSilenceExitCode(1)
		}
		fmt.Fprintf(stderr, "Wrote %s (%d sessions, %d idle, %d busy)\n",
			outPath, snap.Summary.Sessions, snap.Summary.Idle, snap.Summary.Busy)
		return nil
	}

	if err := RenderSnapshot(stdout, snap, opt); err != nil {
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	return nil
}

// runSessionsSnapshotStream emits each CLI window block as soon as that window
// is collected and process-enriched; summary footer is written last.
func runSessionsSnapshotStream(stdout, stderr io.Writer, noColor, noTree bool, capOpts CaptureOpts) error {
	opt := RenderOptions{Format: FormatCLI, NoColor: noColor, NoTree: noTree}
	color := opt.useColor()
	snap, warnings, err := CaptureSnapshotStream(capOpts, func(win SnapshotWindow) error {
		return renderCLIWindow(stdout, win, color, noTree)
	})
	if err != nil {
		WriteError(stderr, strings.TrimPrefix(err.Error(), "Error: "))
		return errs.NewSilenceExitCode(1)
	}
	for _, w := range warnings {
		WriteWarning(stderr, strings.TrimPrefix(w, "warning: "))
	}
	// Window blocks already streamed; footer summary last.
	if err := renderCLIFooter(stdout, snap, color); err != nil {
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	return nil
}

func runSession(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stdout, strings.TrimSpace(sessionHelp)+"\n")
		return nil
	}
	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Fprint(stdout, strings.TrimSpace(sessionHelp)+"\n")
		return nil
	}

	sessionRef := args[0]
	rest := args[1:]
	if len(rest) == 0 {
		fmt.Fprint(stderr, "Error: session: missing command (expected status)\n\n"+strings.TrimSpace(sessionHelp)+"\n")
		return errs.NewSilenceExitCode(1)
	}
	cmd := rest[0]
	cmdArgs := rest[1:]
	switch cmd {
	case "status":
		return runSessionStatus(sessionRef, cmdArgs, stdout, stderr)
	case "-h", "--help", "help":
		fmt.Fprint(stdout, strings.TrimSpace(sessionHelp)+"\n")
		return nil
	default:
		fmt.Fprintf(stderr, "Error: session: unknown command %q\n\n%s\n", cmd, strings.TrimSpace(sessionHelp))
		return errs.NewSilenceExitCode(1)
	}
}

func runSessionStatus(sessionRef string, args []string, stdout, stderr io.Writer) error {
	var asJSON, asMD, asHTML, noColor, noEnrich, noTree bool
	var outPath string
	remain, err := lessflags.Bool("--json", &asJSON).
		Bool("--markdown", &asMD).
		Bool("--html", &asHTML).
		Bool("--no-color", &noColor).
		Bool("--no-enrich", &noEnrich).
		Bool("--no-tree", &noTree).
		String("-o,--output", &outPath).
		HelpFunc("-h,--help", func() {}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			fmt.Fprint(stdout, strings.TrimSpace(sessionHelp)+"\n")
			return nil
		}
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) > 0 {
		fmt.Fprintf(stderr, "Error: session status: unexpected arguments: %s\n", strings.Join(remain, " "))
		return errs.NewSilenceExitCode(1)
	}

	format, err := ResolveFormat(FormatFlags{JSON: asJSON, Markdown: asMD, HTML: asHTML}, outPath)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}

	snap, warnings, err := CaptureSnapshotWith(CaptureOpts{NoEnrich: noEnrich})
	if err != nil {
		WriteError(stderr, strings.TrimPrefix(err.Error(), "Error: "))
		return errs.NewSilenceExitCode(1)
	}
	for _, w := range warnings {
		WriteWarning(stderr, strings.TrimPrefix(w, "warning: "))
	}

	matches := FindSessionsByRef(snap, sessionRef)
	if len(matches) == 0 {
		WriteError(stderr, fmt.Sprintf("session not found: %s", sessionRef))
		return errs.NewSilenceExitCode(1)
	}
	if len(matches) > 1 {
		WriteError(stderr, fmt.Sprintf("ambiguous session id %q (matched %d); use full unique id", sessionRef, len(matches)))
		return errs.NewSilenceExitCode(1)
	}
	s := matches[0]

	opt := RenderOptions{Format: format, NoColor: noColor, NoTree: noTree}
	if outPath != "" {
		f, err := os.Create(outPath)
		if err != nil {
			WriteError(stderr, err.Error())
			return errs.NewSilenceExitCode(1)
		}
		defer f.Close()
		opt.NoColor = true
		if err := RenderSessionStatus(f, snap, s, opt); err != nil {
			WriteError(stderr, err.Error())
			return errs.NewSilenceExitCode(1)
		}
		fmt.Fprintf(stderr, "Wrote %s (session %s)\n", outPath, shortID(s.ID))
		return nil
	}
	if err := RenderSessionStatus(stdout, snap, s, opt); err != nil {
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	return nil
}
