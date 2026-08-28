package iterm2

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	lib "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/tabselect"
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

Usage: kool iterm2 sessions restore [--dry-run] [--force] [--file PATH] [--color|--no-color] [--ignore-macos-space] [--same-app]

Read the checkpoint, create one window per saved window, one tab per entry,
then send: cd <cwd> and grok --resume / codex resume / mark <message>.

  --dry-run              print plan only; do not create tabs or mark restored
  --force                allow a checkpoint with restored_at; live-session
                         duplicate checks still apply
  --file PATH            checkpoint path (default: ~/.config/iterm2/sessions-save.json)
  --color                force ANSI colors on (wins over NO_COLOR / non-TTY)
  --no-color             force ANSI colors off
  --ignore-macos-space   ignore recorded space; create on current Desktop
  --same-app             recreate each window in its recorded app (canonical
                         home/system path). Default prefers
                         ~/Applications/iTerm.app when multiple installs exist;
                         else the only install on disk; else bare iTerm2
  -h, --help

If restored_at is set, the file is already consumed and restore errors unless
--force is supplied. --force does not bypass the live-session duplicate check.
Before creating tabs, restore scans every running iTerm install and skips matching
live grok/codex sessions or mark messages. A live restore aborts if this safety
scan fails; dry-run warns and shows the unfiltered saved layout.
On full success, restored_at is written so the checkpoint cannot be applied twice.
By default each window is placed on its recorded macOS Space (Switch/Create).
If already on that Desktop, Switch is skipped. If Switch fails after retries
(Mission Control AX flake), restore warns and continues on the current Desktop
instead of aborting.
Create target: by default one global install (prefer home when both exist);
--same-app uses each window's recorded app.

Examples:
  kool iterm2 sessions restore --dry-run
  kool iterm2 sessions restore
  kool iterm2 sessions restore --file ~/Desktop/pre-reboot.json
  kool iterm2 sessions restore --ignore-macos-space
  kool iterm2 sessions restore --same-app
`

const sessionHelp = `iterm2 session — inspect or drive a single iTerm2 session

Usage:
  kool iterm2 session <session-id> status [options]
  kool iterm2 session <session-id> send [--focus] [--no-submit] [--no-ctrl-u] <text>
  kool iterm2 session send (--session-id ID | --tab SEL | --tab-index N) [options] <text>
  kool iterm2 session -h|--help

Session id (positional form):
  Primary: iTerm2 session unique ID (UUID from snapshot "id" field).
  Also accepted when unique among live sessions:
    UUID prefix (≥8 chars), tty (ttys003 or /dev/ttys003), or pid.

Commands:
  status                   re-query live status for the session
  send                     type text into the session (AppleScript write text)

Status options:
  --json                   emit JSON
  --markdown               emit Markdown
  --html                   emit HTML
  -o, --output FILE        write to FILE (suffix may infer format)
  --no-color               disable ANSI colors on CLI output
  --no-enrich              skip agent session resolve (no grok/codex session id)
  --no-tree                keep agent session id but omit process tree lines
  -h, --help               show this help

Send options (both forms):
  --focus                  switch to the session's window/tab before writing
  --no-submit              write without newline (stage; user presses Enter)
  --no-ctrl-u              do not prefix Ctrl-U (default prefixes Ctrl-U)
  -h, --help               show this help

Flag-form send sources (exactly one; kool iterm2 session send …):
  --session-id ID          same resolve rules as positional <session-id>
  --tab SEL                1-based tab index, or next|left|right (right ≡ next)
  --tab-index N            0-based tab index in this iTerm window
  Tab selectors use the same window/tab discovery as: kool iterm2 window status.
  --tab / --tab-index / --session-id are not valid on the positional form.

Examples:
  kool iterm2 session D922B298-25FB-41FA-BAF8-7AC7A1D56758 status
  kool iterm2 session D922B298 status --json
  kool iterm2 session ttys003 status
  kool iterm2 session D922B298 send "echo hi"
  kool iterm2 session D922B298 send --no-submit --no-ctrl-u "partial"
  kool iterm2 session D922B298 send --focus "ls"
  kool iterm2 session send --tab next "echo hi"
  kool iterm2 session send --tab 2 --focus "ls"
  kool iterm2 session send --tab-index 0 "echo hi"
  kool iterm2 session send --session-id D922B298 "echo hi"
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

func runSession(args []string, stdout, stderr io.Writer, env TestRun) error {
	if len(args) == 0 {
		fmt.Fprint(stdout, strings.TrimSpace(sessionHelp)+"\n")
		return nil
	}
	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Fprint(stdout, strings.TrimSpace(sessionHelp)+"\n")
		return nil
	}

	// Flag form: kool iterm2 session send (--session-id|… ) <text>
	if args[0] == "send" {
		return runSessionSendFlags(args[1:], stdout, stderr, env)
	}

	sessionRef := args[0]
	rest := args[1:]
	if len(rest) == 0 {
		fmt.Fprint(stderr, "Error: session: missing command (expected status or send)\n\n"+strings.TrimSpace(sessionHelp)+"\n")
		return errs.NewSilenceExitCode(1)
	}
	cmd := rest[0]
	cmdArgs := rest[1:]
	switch cmd {
	case "status":
		return runSessionStatus(sessionRef, cmdArgs, stdout, stderr)
	case "send":
		return runSessionSendPositional(sessionRef, cmdArgs, stdout, stderr, env)
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

// runSessionSendPositional handles: session <session-id> send [flags] <text>
func runSessionSendPositional(sessionRef string, args []string, stdout, stderr io.Writer, env TestRun) error {
	var focus, noSubmit, noCtrlU bool
	var sessionIDFlag, tabFlag string
	var tabIndex *int
	remain, err := lessflags.Bool("--focus", &focus).
		Bool("--no-submit", &noSubmit).
		Bool("--no-ctrl-u", &noCtrlU).
		String("--session-id", &sessionIDFlag).
		String("--tab", &tabFlag).
		Int("--tab-index", &tabIndex).
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
	if sessionIDFlag != "" || tabFlag != "" || tabIndex != nil {
		WriteError(stderr, "session send: --session-id / --tab / --tab-index belong on: kool iterm2 session send …")
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) == 0 {
		WriteError(stderr, "session send: missing text")
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) > 1 {
		WriteError(stderr, fmt.Sprintf("session send: unexpected arguments: %s", strings.Join(remain[1:], " ")))
		return errs.NewSilenceExitCode(1)
	}
	text := remain[0]

	targetID, err := resolveSendSessionID(sessionRef, env)
	if err != nil {
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	return finishSessionSend(sessionRef, targetID, text, lib.SendTextOptions{
		Focus:    focus,
		NoSubmit: noSubmit,
		NoCtrlU:  noCtrlU,
	}, stdout, stderr, env)
}

// runSessionSendFlags handles: session send (--session-id|…|--) [flags] <text>
func runSessionSendFlags(args []string, stdout, stderr io.Writer, env TestRun) error {
	var focus, noSubmit, noCtrlU bool
	var sessionIDFlag, tabFlag string
	var tabIndex *int
	remain, err := lessflags.Bool("--focus", &focus).
		Bool("--no-submit", &noSubmit).
		Bool("--no-ctrl-u", &noCtrlU).
		String("--session-id", &sessionIDFlag).
		String("--tab", &tabFlag).
		Int("--tab-index", &tabIndex).
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
	if len(remain) == 0 {
		WriteError(stderr, "session send: missing text")
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) > 1 {
		WriteError(stderr, fmt.Sprintf("session send: unexpected arguments: %s", strings.Join(remain[1:], " ")))
		return errs.NewSilenceExitCode(1)
	}
	text := remain[0]

	hasSID := strings.TrimSpace(sessionIDFlag) != ""
	hasTab := strings.TrimSpace(tabFlag) != ""
	hasTabIndex := tabIndex != nil
	nSources := 0
	if hasSID {
		nSources++
	}
	if hasTab {
		nSources++
	}
	if hasTabIndex {
		nSources++
	}
	if nSources == 0 {
		WriteError(stderr, "session send: expected --session-id, or --tab / --tab-index")
		return errs.NewSilenceExitCode(1)
	}
	if nSources > 1 {
		if hasSID && (hasTab || hasTabIndex) {
			WriteError(stderr, "--session-id cannot be combined with --tab/--tab-index")
			return errs.NewSilenceExitCode(1)
		}
		WriteError(stderr, "--tab and --tab-index cannot be specified together")
		return errs.NewSilenceExitCode(1)
	}

	opts := lib.SendTextOptions{
		Focus:    focus,
		NoSubmit: noSubmit,
		NoCtrlU:  noCtrlU,
	}

	if hasSID {
		ref := strings.TrimSpace(sessionIDFlag)
		targetID, err := resolveSendSessionID(ref, env)
		if err != nil {
			WriteError(stderr, err.Error())
			return errs.NewSilenceExitCode(1)
		}
		return finishSessionSend(ref, targetID, text, opts, stdout, stderr, env)
	}

	var sel tabselect.TabSelector
	if hasTab {
		sel, err = tabselect.ParseTabFlag(tabFlag)
	} else {
		sel, err = tabselect.ParseTabIndexFlag(*tabIndex)
	}
	if err != nil {
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}

	targetID, display, err := resolveSendTabSessionID(sel, env)
	if err != nil {
		if errors.Is(err, lib.ErrNotInSession) {
			WriteError(stderr, "iterm2: not inside an iTerm2 session (no ITERM_SESSION_ID and no matching TTY)")
			return errs.NewSilenceExitCode(1)
		}
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	return finishSessionSend(display, targetID, text, opts, stdout, stderr, env)
}

func finishSessionSend(displayRef, targetID, text string, opts lib.SendTextOptions, stdout, stderr io.Writer, env TestRun) error {
	sendFn := env.SendText
	if sendFn == nil {
		sendFn = lib.SendText
	}
	if err := sendFn(targetID, text, opts, nil); err != nil {
		msg := err.Error()
		if errors.Is(err, lib.ErrSessionNotFound) || strings.Contains(strings.ToLower(msg), "session not found") {
			ref := strings.TrimSpace(displayRef)
			if ref == "" {
				ref = targetID
			}
			WriteError(stderr, fmt.Sprintf("session not found: %s", ref))
			return errs.NewSilenceExitCode(1)
		}
		WriteError(stderr, strings.TrimPrefix(msg, "Error: "))
		return errs.NewSilenceExitCode(1)
	}

	display := strings.TrimSpace(displayRef)
	if display == "" {
		display = shortID(targetID)
	}
	fmt.Fprintf(stdout, "sent to session %s\n", display)
	return nil
}

// resolveSendTabSessionID maps a tab selector in the current window to an iTerm session UUID.
func resolveSendTabSessionID(sel tabselect.TabSelector, env TestRun) (targetID, display string, err error) {
	st, err := lib.CurrentWindowStatusWith(env.currentStatusConfig())
	if err != nil {
		return "", "", err
	}
	row, _, err := tabselect.SelectWindowTab(st, sel)
	if err != nil {
		return "", "", err
	}
	id := strings.TrimSpace(row.SessionID)
	if id == "" {
		return "", "", fmt.Errorf("no session id on tab %d", row.Index)
	}
	uuid := lib.SessionUUID(id)
	if uuid == "" {
		uuid = id
	}
	return uuid, uuid, nil
}

// resolveSendSessionID maps a user session ref to a full unique ID for SendText.
// Full UUIDs skip any live scan (fast path). Prefix/tty use light ListSessions.
// Numeric PID falls back to CaptureSnapshot (rare; ListSessions has no PIDs).
func resolveSendSessionID(sessionRef string, env TestRun) (string, error) {
	ref := strings.TrimSpace(sessionRef)
	if ref == "" {
		return "", errors.New("session send: missing session-id")
	}
	uuid := lib.SessionUUID(ref)
	if lib.IsFullSessionUUID(uuid) {
		return uuid, nil
	}

	listFn := env.ListSessions
	if listFn == nil && env.CurrentStatus != nil && env.CurrentStatus.ListSessions != nil {
		listFn = env.CurrentStatus.ListSessions
	}
	if listFn == nil {
		listFn = lib.ListSessions
	}
	refs, err := listFn()
	if err != nil {
		return "", errors.New(strings.TrimPrefix(err.Error(), "Error: "))
	}
	matches := lib.FindSessionRefsByRef(refs, ref)
	if len(matches) == 1 {
		id := strings.TrimSpace(matches[0].SessionID)
		if id == "" {
			return "", fmt.Errorf("session not found: %s", ref)
		}
		return id, nil
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("ambiguous session id %q (matched %d); use full unique id", ref, len(matches))
	}

	// PID-only refs need the heavy snapshot (session list has no PIDs).
	if _, err := strconv.Atoi(ref); err == nil {
		snap, _, err := CaptureSnapshotWith(CaptureOpts{NoEnrich: true})
		if err != nil {
			return "", errors.New(strings.TrimPrefix(err.Error(), "Error: "))
		}
		sm := FindSessionsByRef(snap, ref)
		if len(sm) == 0 {
			return "", fmt.Errorf("session not found: %s", ref)
		}
		if len(sm) > 1 {
			return "", fmt.Errorf("ambiguous session id %q (matched %d); use full unique id", ref, len(sm))
		}
		return sm[0].ID, nil
	}

	return "", fmt.Errorf("session not found: %s", ref)
}
