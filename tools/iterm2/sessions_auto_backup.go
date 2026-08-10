package iterm2

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/kool/pkgs/duration"
	"github.com/xhd2015/kool/pkgs/errs"
)

const (
	sessionsAutoSource     = "kool-iterm2-sessions-auto"
	defaultAutoIntervalStr = "10m"
)

// DefaultSessionsAutoPath is ~/.config/iterm2/sessions-auto.json.
// Distinct from DefaultSessionsSavePath (manual sessions-save.json).
func DefaultSessionsAutoPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = "."
	}
	return filepath.Join(home, ".config", "iterm2", "sessions-auto.json")
}

// Optional test override when --file is omitted.
var sessionsAutoPathForTest string

// effectiveAutoPath returns --file or the auto default (test override aware).
func effectiveAutoPath(fileFlag string) string {
	if fileFlag != "" {
		return fileFlag
	}
	if sessionsAutoPathForTest != "" {
		return sessionsAutoPathForTest
	}
	return DefaultSessionsAutoPath()
}

const sessionsAutoBackupHelp = `iterm2 sessions auto-backup — periodically checkpoint critical tabs for crash recovery

Usage: kool iterm2 sessions auto-backup [options]

Periodically checkpoint critical grok/codex/mark tabs (same filter as sessions
save) so crash recovery can restore from the auto file. Default interval is
10m; the first cycle runs immediately, then sleeps --interval.

Default checkpoint: ~/.config/iterm2/sessions-auto.json
(distinct from manual sessions-save.json). Always overwrites the auto file on
a successful non-empty write (no TTY prompt). Capture / iTerm failures soft-warn
and keep the previous backup. Zero critical sessions: message only, no write.

  --interval DUR         sleep between cycles (default 10m); accepts 10m, 60s,
                         or bare seconds (pkgs/duration.Parse)
  --file PATH            checkpoint path (default: ~/.config/iterm2/sessions-auto.json)
  --once                 run a single cycle and exit
  --dry-run              plan only; do not write
  --color                force ANSI colors on (wins over NO_COLOR / non-TTY)
  --no-color             force ANSI colors off
  --ignore-macos-space   omit space / iterm_window_id; do not resolve Spaces
  --spaces LIST          only save windows on these 0-based Space indexes
                         (comma-separated, e.g. 0,2); cannot combine with
                         --ignore-macos-space
  -h, --help

Restore with: kool iterm2 sessions restore --file ~/.config/iterm2/sessions-auto.json

Examples:
  kool iterm2 sessions auto-backup
  kool iterm2 sessions auto-backup --once
  kool iterm2 sessions auto-backup --interval 5m --once
  kool iterm2 sessions auto-backup --once --dry-run
  kool iterm2 sessions auto-backup --once --file ~/Desktop/crash-auto.json
`

// runSessionsAutoBackup implements `sessions auto-backup`.
func runSessionsAutoBackup(args []string, stdout, stderr io.Writer) error {
	var (
		dryRun, once             bool
		fileFlag, intervalRaw    string
		forceColor, forceNoColor bool
		ignoreMacOSSpace         bool
		spacesRaw                string
		spacesSet                bool
	)

	remain, err := parseAutoBackupFlags(args, &dryRun, &once, &fileFlag, &intervalRaw,
		&forceColor, &forceNoColor, &ignoreMacOSSpace, &spacesRaw, &spacesSet)
	if err != nil {
		if err == errHelpRequested {
			fmt.Fprint(stdout, strings.TrimSpace(sessionsAutoBackupHelp)+"\n")
			return nil
		}
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) > 0 {
		fmt.Fprintf(stderr, "Error: sessions auto-backup: unexpected arguments: %s\n", strings.Join(remain, " "))
		return errs.NewSilenceExitCode(1)
	}
	if spacesSet && ignoreMacOSSpace {
		WriteError(stderr, "--spaces cannot be used with --ignore-macos-space")
		return errs.NewSilenceExitCode(1)
	}

	// Interval: default 10m; invalid → Error before loop.
	if intervalRaw == "" {
		intervalRaw = defaultAutoIntervalStr
	}
	interval, err := duration.Parse(intervalRaw)
	if err != nil {
		WriteError(stderr, fmt.Sprintf("invalid --interval %q: %v", intervalRaw, err))
		return errs.NewSilenceExitCode(1)
	}

	var spacesAllow []int
	if spacesSet {
		spacesAllow, err = parseSpacesList(spacesRaw)
		if err != nil {
			WriteError(stderr, err.Error())
			return errs.NewSilenceExitCode(1)
		}
	}

	color, err := resolvePlanColor(forceColor, forceNoColor)
	if err != nil {
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}

	path := effectiveAutoPath(fileFlag)

	// First tick immediate; --once exits after one cycle.
	for {
		if err := runAutoBackupCycle(stdout, stderr, path, dryRun, color, ignoreMacOSSpace, spacesAllow); err != nil {
			// Hard errors (e.g. write failure) propagate; capture soft-fails inside cycle.
			return err
		}
		if once {
			return nil
		}
		time.Sleep(interval)
	}
}

// runAutoBackupCycle performs one capture → filter → write/plan cycle.
// Capture failures soft-warn and return nil (exit 0 for --once).
// Zero critical: message + no write. Non-empty: always overwrite (no TTY prompt).
func runAutoBackupCycle(stdout, stderr io.Writer, path string, dryRun bool, color bool, ignoreMacOSSpace bool, spacesAllow []int) error {
	spaceSkipped := 0
	capOpts := CaptureOpts{NoEnrich: false, SpaceAllow: spacesAllow, SpaceSkipped: &spaceSkipped}
	snap, warnings, err := CaptureSnapshotForSave(capOpts)
	if err != nil {
		// Soft-fail: keep previous backup; cycle success (exit 0 with --once).
		msg := strings.TrimPrefix(err.Error(), "Error: ")
		msg = strings.TrimSpace(msg)
		if msg == "" {
			msg = "snapshot capture failed"
		}
		WriteWarning(stderr, msg)
		return nil
	}
	for _, w := range warnings {
		WriteWarning(stderr, strings.TrimPrefix(w, "warning: "))
	}

	now := sessionsNowFn()
	host, _ := sessionsHostnameFn()
	doc, filterWarns := BuildSaveDocument(snap, now, host)
	for _, w := range filterWarns {
		WriteWarning(stderr, strings.TrimPrefix(w, "warning: "))
	}
	for _, w := range applySpaceToSaveDocument(doc, snap, ignoreMacOSSpace) {
		WriteWarning(stderr, strings.TrimPrefix(w, "warning: "))
	}

	// Multi-app: attach canonical app, hard-dedupe, dual-collapse warn.
	pf := resolveMultiAppPreflight()
	var collapseWarn string
	doc.Windows, collapseWarn = applyMultiAppToSaveWindows(doc.Windows, snap, pf)
	recomputeSaveSummary(doc)
	if collapseWarn != "" {
		WriteWarning(stderr, collapseWarn)
	}

	skippedSpaces := spaceSkipped
	if len(spacesAllow) > 0 {
		doc.Filter = &SaveFilter{Spaces: append([]int(nil), spacesAllow...)}
		if n := filterSaveDocumentBySpaces(doc, spacesAllow); n > 0 {
			skippedSpaces += n
		}
	}
	sortSaveWindowsBySpace(doc.Windows)
	recomputeSaveSummary(doc)

	// Auto checkpoint provenance (distinct from manual save).
	doc.Source = sessionsAutoSource
	doc.RestoredAt = nil

	if doc.Summary.Sessions == 0 {
		fmt.Fprintln(stdout, "0 critical sessions (nothing to save; previous backup kept)")
		if skippedSpaces > 0 {
			WriteWarning(stderr, formatSpacesSkipWarning(skippedSpaces, spacesAllow))
		}
		return nil
	}

	if dryRun {
		formatSavePlan(stdout, doc, path, true, color)
		if skippedSpaces > 0 {
			WriteWarning(stderr, formatSpacesSkipWarning(skippedSpaces, spacesAllow))
		}
		return nil
	}

	// Always overwrite on successful non-empty write (no TTY pending prompt).
	if err := WriteSaveDocument(path, doc); err != nil {
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	formatSavePlan(stdout, doc, path, false, color)
	if skippedSpaces > 0 {
		WriteWarning(stderr, formatSpacesSkipWarning(skippedSpaces, spacesAllow))
	}
	return nil
}

// parseAutoBackupFlags parses auto-backup CLI flags.
func parseAutoBackupFlags(args []string, dryRun, once *bool, fileFlag, intervalRaw *string,
	forceColor, forceNoColor *bool, ignoreMacOSSpace *bool, spacesRaw *string, spacesSet *bool) ([]string, error) {
	var remain []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help" || a == "help":
			return nil, errHelpRequested
		case a == "--dry-run":
			*dryRun = true
		case a == "--once":
			*once = true
		case a == "--color":
			*forceColor = true
		case a == "--no-color":
			*forceNoColor = true
		case a == "--ignore-macos-space":
			*ignoreMacOSSpace = true
		case a == "--spaces":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--spaces requires a comma-separated list of space indexes (e.g. 0,2)")
			}
			i++
			*spacesSet = true
			*spacesRaw = args[i]
		case strings.HasPrefix(a, "--spaces="):
			*spacesSet = true
			*spacesRaw = strings.TrimPrefix(a, "--spaces=")
		case a == "--file" || a == "-f":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--file requires a path")
			}
			i++
			*fileFlag = args[i]
		case strings.HasPrefix(a, "--file="):
			*fileFlag = strings.TrimPrefix(a, "--file=")
		case a == "--interval":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--interval requires a duration (e.g. 10m, 60s)")
			}
			i++
			*intervalRaw = args[i]
		case strings.HasPrefix(a, "--interval="):
			*intervalRaw = strings.TrimPrefix(a, "--interval=")
		default:
			if strings.HasPrefix(a, "-") {
				return nil, fmt.Errorf("unknown flag %s", a)
			}
			remain = append(remain, a)
		}
	}
	return remain, nil
}
