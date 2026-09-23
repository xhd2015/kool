package iterm2

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/kool/pkgs/errs"
	lessflags "github.com/xhd2015/less-flags"
	"golang.org/x/term"
)

const sessionListHelp = `iterm2 session list — list live iTerm2 sessions (flat)

Usage: kool iterm2 session list [options]

Options:
  --grok                   only sessions with a resolved grok agent
  --only-cwd               only sessions whose cwd equals process cwd
  --json                   emit JSON array (no ANSI)
  --no-color               disable ANSI colors on CLI output
  --no-enrich              skip agent session resolve (no kind / agent id)
  -h, --help               show this help

Default enrich is on so KIND and AGENT SESSION columns are useful.
--only-cwd compares realpaths (Abs + EvalSymlinks).

Examples:
  kool iterm2 session list
  kool iterm2 session list --grok --only-cwd
  kool iterm2 session list --json
`

// SessionListRow is one flat row for session list output.
type SessionListRow struct {
	ItermID      string `json:"iterm_id"`
	Status       string `json:"status"` // idle | busy | unknown
	Name         string `json:"name,omitempty"`
	Cwd          string `json:"cwd,omitempty"`
	Kind         string `json:"kind,omitempty"` // grok | codex | …
	AgentSession string `json:"agent_session,omitempty"`
	Title        string `json:"title,omitempty"`
	WindowIndex  int    `json:"window_index,omitempty"`
	TabIndex     int    `json:"tab_index,omitempty"`
	PaneIndex    int    `json:"pane_index,omitempty"`
}

// FlattenSessions builds list rows from a full snapshot (window order preserved).
func FlattenSessions(snap *Snapshot) []SessionListRow {
	if snap == nil {
		return nil
	}
	var rows []SessionListRow
	for _, win := range snap.Windows {
		for _, tab := range win.Tabs {
			for _, s := range tab.Sessions {
				row := SessionListRow{
					ItermID:     s.ID,
					Status:      idleLabel(s.Idle),
					Name:        firstNonEmpty(s.Name, tab.Name),
					Cwd:         derefStr(s.Cwd),
					WindowIndex: win.Index,
					TabIndex:    tab.Index,
					PaneIndex:   s.Index,
				}
				if s.Agent != nil {
					row.Kind = strings.ToLower(strings.TrimSpace(s.Agent.Kind))
					row.AgentSession = s.Agent.SessionID
					row.Title = s.Agent.Title
				}
				rows = append(rows, row)
			}
		}
	}
	return rows
}

// FilterSessionList applies --grok and/or --only-cwd filters.
// onlyCwdPath should be the caller's realpath cwd when onlyCwd is true; empty
// onlyCwdPath with onlyCwd true yields no matches (safe).
func FilterSessionList(rows []SessionListRow, grokOnly, onlyCwd bool, onlyCwdPath string) []SessionListRow {
	if !grokOnly && !onlyCwd {
		return rows
	}
	wantCwd := ""
	if onlyCwd {
		wantCwd = normalizePathForCompare(onlyCwdPath)
	}
	out := make([]SessionListRow, 0, len(rows))
	for _, r := range rows {
		if grokOnly {
			if r.Kind != "grok" || strings.TrimSpace(r.AgentSession) == "" {
				continue
			}
		}
		if onlyCwd {
			if wantCwd == "" || normalizePathForCompare(r.Cwd) != wantCwd {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

// normalizePathForCompare returns Abs+EvalSymlinks form for equality checks.
// Empty input stays empty. On error, falls back to cleaned Abs or original.
func normalizePathForCompare(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return filepath.Clean(p)
	}
	if real, err := filepath.EvalSymlinks(abs); err == nil {
		return real
	}
	return abs
}

// ProcessCwd returns the process working directory, normalized for --only-cwd.
func ProcessCwd() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return normalizePathForCompare(wd), nil
}

func runSessionList(args []string, stdout, stderr io.Writer) error {
	var grokOnly, onlyCwd, asJSON, noColor, noEnrich bool
	remain, err := lessflags.Bool("--grok", &grokOnly).
		Bool("--only-cwd", &onlyCwd).
		Bool("--json", &asJSON).
		Bool("--no-color", &noColor).
		Bool("--no-enrich", &noEnrich).
		HelpFunc("-h,--help", func() {}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			fmt.Fprint(stdout, strings.TrimSpace(sessionListHelp)+"\n")
			return nil
		}
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) > 0 {
		fmt.Fprintf(stderr, "Error: session list: unexpected arguments: %s\n", strings.Join(remain, " "))
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

	rows := FlattenSessions(snap)
	cwdPath := ""
	if onlyCwd {
		cwdPath, err = ProcessCwd()
		if err != nil {
			WriteError(stderr, fmt.Sprintf("session list: cwd: %v", err))
			return errs.NewSilenceExitCode(1)
		}
	}
	rows = FilterSessionList(rows, grokOnly, onlyCwd, cwdPath)

	if asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		if err := enc.Encode(rows); err != nil {
			WriteError(stderr, err.Error())
			return errs.NewSilenceExitCode(1)
		}
		return nil
	}

	color := !noColor && useCLIColor(stdout)
	return renderSessionListCLI(stdout, rows, color)
}

// useCLIColor mirrors RenderOptions.useColor for list (CLI only, no ForceColor).
func useCLIColor(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func renderSessionListCLI(w io.Writer, rows []SessionListRow, color bool) error {
	if len(rows) == 0 {
		fmt.Fprintln(w, paint(color, ansiGray, "0 sessions"))
		return nil
	}

	// Column widths (minimums from headers).
	idW, stW, kindW, agentW, cwdW := 2, 6, 4, 13, 3
	type line struct {
		id, status, kind, agent, cwd string
	}
	lines := make([]line, 0, len(rows))
	for _, r := range rows {
		id := shortID(r.ItermID)
		if id == "" {
			id = "-"
		}
		st := r.Status
		if st == "" {
			st = "unknown"
		}
		kind := r.Kind
		if kind == "" {
			kind = "-"
		}
		agent := r.AgentSession
		if agent == "" {
			agent = "-"
		}
		cwd := r.Cwd
		if cwd == "" {
			cwd = "-"
		}
		if len(id) > idW {
			idW = len(id)
		}
		if len(st) > stW {
			stW = len(st)
		}
		if len(kind) > kindW {
			kindW = len(kind)
		}
		if len(agent) > agentW {
			agentW = len(agent)
		}
		if len(cwd) > cwdW {
			cwdW = len(cwd)
		}
		lines = append(lines, line{id: id, status: st, kind: kind, agent: agent, cwd: cwd})
	}

	hdr := fmt.Sprintf("  %-*s  %-*s  %-*s  %-*s  %-*s",
		idW, "ID", stW, "STATUS", kindW, "KIND", agentW, "AGENT SESSION", cwdW, "CWD")
	fmt.Fprintln(w, paint(color, ansiGray, hdr))

	for _, ln := range lines {
		st := ln.status
		switch ln.status {
		case "idle":
			st = paint(color, ansiGreen, ln.status)
		case "busy":
			st = paint(color, ansiYellow, ln.status)
		default:
			st = paint(color, ansiGray, ln.status)
		}
		// Pad status using plain width (ANSI-safe enough for fixed layout).
		padStatus := ln.status
		if len(padStatus) < stW {
			// reprint with visual pad after colored token
			st = st + strings.Repeat(" ", stW-len(padStatus))
		}
		fmt.Fprintf(w, "  %-*s  %s  %-*s  %-*s  %s\n",
			idW, ln.id, st, kindW, ln.kind, agentW, ln.agent, ln.cwd)
	}

	n := len(rows)
	label := "sessions"
	if n == 1 {
		label = "session"
	}
	fmt.Fprintln(w, paint(color, ansiGray, fmt.Sprintf("%d %s", n, label)))
	return nil
}
