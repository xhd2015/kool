package iterm2

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// color helpers (ANSI). Disabled when !useColor.
const (
	ansiReset  = "\x1b[0m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
	ansiGray   = "\x1b[90m"
	ansiBold   = "\x1b[1m"
	ansiRed    = "\x1b[31m"
)

// RenderOptions control snapshot rendering.
type RenderOptions struct {
	Format  OutputFormat
	NoColor bool
	// NoTree omits FormatTree process-tree lines while keeping agent session id.
	NoTree bool
	// ForceColor forces ANSI even when stdout is not a TTY (tests).
	ForceColor bool
	// ColorWriter is checked for TTY (defaults to os.Stdout).
	ColorWriter *os.File
}

func (o RenderOptions) useColor() bool {
	if o.Format != FormatCLI {
		return false
	}
	if o.NoColor {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if o.ForceColor {
		return true
	}
	f := o.ColorWriter
	if f == nil {
		f = os.Stdout
	}
	return term.IsTerminal(int(f.Fd()))
}

func paint(use bool, code, s string) string {
	if !use {
		return s
	}
	return code + s + ansiReset
}

// RenderSnapshot writes the snapshot in the requested format to w.
func RenderSnapshot(w io.Writer, snap *Snapshot, opt RenderOptions) error {
	switch opt.Format {
	case FormatJSON:
		return renderJSON(w, snap)
	case FormatMarkdown:
		return renderMarkdown(w, snap)
	case FormatHTML:
		return renderHTML(w, snap)
	default:
		return renderCLI(w, snap, opt.useColor(), opt.NoTree)
	}
}

// RenderSessionStatus writes a single-session status report.
func RenderSessionStatus(w io.Writer, snap *Snapshot, s *SnapshotSession, opt RenderOptions) error {
	switch opt.Format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		return enc.Encode(s)
	case FormatMarkdown:
		return renderSessionMarkdown(w, snap, s)
	case FormatHTML:
		return renderSessionHTML(w, snap, s)
	default:
		return renderSessionCLI(w, snap, s, opt.useColor(), opt.NoTree)
	}
}

func renderJSON(w io.Writer, snap *Snapshot) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(snap)
}

func idleLabel(idle *bool) string {
	if idle == nil {
		return "unknown"
	}
	if *idle {
		return "idle"
	}
	return "busy"
}

func shortID(id string) string {
	id = strings.TrimSpace(id)
	if len(id) <= 8 {
		return id
	}
	// prefer first 8 of UUID without relying on dashes
	compact := strings.ReplaceAll(id, "-", "")
	if len(compact) >= 8 {
		return compact[:8]
	}
	return id[:8]
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func derefInt(p *int) string {
	if p == nil {
		return "-"
	}
	return fmt.Sprintf("%d", *p)
}

func renderCLI(w io.Writer, snap *Snapshot, color bool, noTree bool) error {
	if err := renderCLIHeader(w, snap, color); err != nil {
		return err
	}
	for _, win := range snap.Windows {
		if err := renderCLIWindow(w, win, color, noTree); err != nil {
			return err
		}
	}
	return renderCLIFooter(w, snap, color)
}

func renderCLIHeader(w io.Writer, snap *Snapshot, color bool) error {
	gray := func(s string) string { return paint(color, ansiGray, s) }
	bold := func(s string) string { return paint(color, ansiBold, s) }
	fmt.Fprintf(w, "%s  %d windows  %d tabs  %d sessions  idle %d  busy %d",
		bold("iTerm2 snapshot"),
		snap.Summary.Windows, snap.Summary.Tabs, snap.Summary.Sessions,
		snap.Summary.Idle, snap.Summary.Busy)
	if snap.Summary.Unknown > 0 {
		fmt.Fprintf(w, "  unknown %d", snap.Summary.Unknown)
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s\n\n", gray(fmt.Sprintf("captured %s  host %s", snap.CapturedAt, snap.Host)))
	return nil
}

// renderCLIWindow writes one window block (W{n} … sessions). Used by buffered
// and streaming CLI paths.
func renderCLIWindow(w io.Writer, win SnapshotWindow, color bool, noTree bool) error {
	gray := func(s string) string { return paint(color, ansiGray, s) }
	bold := func(s string) string { return paint(color, ansiBold, s) }
	st := func(idle *bool) string {
		lab := idleLabel(idle)
		switch lab {
		case "idle":
			return paint(color, ansiGreen, lab)
		case "busy":
			return paint(color, ansiYellow, lab)
		default:
			return gray(lab)
		}
	}

	title := win.Name
	if title == "" {
		title = "(untitled)"
	}
	fmt.Fprintf(w, "%s  %s\n", bold(fmt.Sprintf("W%d", win.Index)), title)
	for _, tab := range win.Tabs {
		for _, s := range tab.Sessions {
			dur := derefStr(s.Duration)
			if dur == "" {
				dur = "-"
			}
			cmd := derefStr(s.Command)
			if cmd == "" {
				cmd = "-"
			}
			fmt.Fprintf(w, "  T%d S%d  %s  id %s  pid %s  %s  %s\n",
				tab.Index, s.Index,
				st(s.Idle),
				gray(shortID(s.ID)),
				gray(derefInt(s.PID)),
				gray(dur),
				cmd,
			)
			ttyShort := strings.TrimPrefix(s.TTY, "/dev/")
			cwd := derefStr(s.Cwd)
			if cwd == "" {
				cwd = "-"
			}
			fmt.Fprintf(w, "         %s  cwd %s\n", gray(ttyShort), cwd)
			if s.CommandLine != nil && *s.CommandLine != "" && *s.CommandLine != cmd {
				line := *s.CommandLine
				if len(line) > 100 {
					line = line[:100] + "…"
				}
				fmt.Fprintf(w, "         %s\n", gray(line))
			}
			renderCLIAgent(w, s.Agent, color, noTree)
			// Mark line only when no agent was attached (agent preferred).
			if s.Agent == nil || s.Agent.SessionID == "" {
				if msg, ok := detectMarkSession(&s); ok {
					if msg == "" {
						fmt.Fprintf(w, "         mark\n")
					} else {
						fmt.Fprintf(w, "         mark  %s\n", msg)
					}
				}
			}
		}
	}
	fmt.Fprintln(w)
	return nil
}

// renderCLIAgent writes agent session id + optional title + FormatTree lines.
func renderCLIAgent(w io.Writer, agent *SessionAgent, color bool, noTree bool) {
	if agent == nil || agent.SessionID == "" {
		return
	}
	gray := func(s string) string { return paint(color, ansiGray, s) }
	kind := agent.Kind
	if kind == "" {
		kind = "agent"
	}
	fmt.Fprintf(w, "         %s  session %s\n", kind, agent.SessionID)
	if agent.Title != "" {
		fmt.Fprintf(w, "         %s\n", gray("title "+agent.Title))
	}
	if noTree || len(agent.Tree) == 0 {
		return
	}
	tree := formatAgentTree(agent.Tree)
	if tree == "" {
		return
	}
	for _, line := range strings.Split(tree, "\n") {
		fmt.Fprintf(w, "         %s\n", line)
	}
}

func renderCLIFooter(w io.Writer, snap *Snapshot, color bool) error {
	gray := func(s string) string { return paint(color, ansiGray, s) }
	fmt.Fprintf(w, "%s\n", gray(fmt.Sprintf("%d sessions  (%d idle, %d busy)",
		snap.Summary.Sessions, snap.Summary.Idle, snap.Summary.Busy)))
	return nil
}

func renderSessionCLI(w io.Writer, snap *Snapshot, s *SnapshotSession, color bool, noTree bool) error {
	lab := idleLabel(s.Idle)
	status := lab
	switch lab {
	case "idle":
		status = paint(color, ansiGreen, lab)
	case "busy":
		status = paint(color, ansiYellow, lab)
	}
	fmt.Fprintf(w, "id:          %s\n", s.ID)
	fmt.Fprintf(w, "status:      %s\n", status)
	fmt.Fprintf(w, "window:      %d", s.WindowIndex)
	if snap != nil {
		for _, win := range snap.Windows {
			if win.Index == s.WindowIndex && win.Name != "" {
				fmt.Fprintf(w, "  (%s)", win.Name)
				break
			}
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "tab:         %d\n", s.TabIndex)
	fmt.Fprintf(w, "pane:        %d\n", s.Index)
	fmt.Fprintf(w, "name:        %s\n", s.Name)
	fmt.Fprintf(w, "tty:         %s\n", s.TTY)
	fmt.Fprintf(w, "profile:     %s\n", s.Profile)
	fmt.Fprintf(w, "pid:         %s\n", derefInt(s.PID))
	fmt.Fprintf(w, "shell_pid:   %s\n", derefInt(s.ShellPID))
	fmt.Fprintf(w, "command:     %s\n", derefStr(s.Command))
	fmt.Fprintf(w, "command_line: %s\n", derefStr(s.CommandLine))
	fmt.Fprintf(w, "cwd:         %s\n", derefStr(s.Cwd))
	fmt.Fprintf(w, "start_time:  %s\n", derefStr(s.StartTime))
	fmt.Fprintf(w, "duration:    %s\n", derefStr(s.Duration))
	fmt.Fprintf(w, "iterm_is_processing: %v\n", s.ItermIsProcessing)
	if s.Agent != nil && s.Agent.SessionID != "" {
		kind := s.Agent.Kind
		if kind == "" {
			kind = "agent"
		}
		fmt.Fprintf(w, "agent:       %s\n", kind)
		fmt.Fprintf(w, "session_id:  %s\n", s.Agent.SessionID)
		if s.Agent.Title != "" {
			fmt.Fprintf(w, "title:       %s\n", s.Agent.Title)
		}
		if !noTree && len(s.Agent.Tree) > 0 {
			tree := formatAgentTree(s.Agent.Tree)
			if tree != "" {
				fmt.Fprintln(w, "tree:")
				for _, line := range strings.Split(tree, "\n") {
					fmt.Fprintf(w, "  %s\n", line)
				}
			}
		}
	}
	return nil
}

func renderMarkdown(w io.Writer, snap *Snapshot) error {
	fmt.Fprintf(w, "# iTerm2 snapshot\n\n")
	fmt.Fprintf(w, "- **captured**: %s\n", snap.CapturedAt)
	fmt.Fprintf(w, "- **host**: %s\n", snap.Host)
	fmt.Fprintf(w, "- **windows**: %d · **tabs**: %d · **sessions**: %d\n",
		snap.Summary.Windows, snap.Summary.Tabs, snap.Summary.Sessions)
	fmt.Fprintf(w, "- **idle**: %d · **busy**: %d", snap.Summary.Idle, snap.Summary.Busy)
	if snap.Summary.Unknown > 0 {
		fmt.Fprintf(w, " · **unknown**: %d", snap.Summary.Unknown)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w)

	for _, win := range snap.Windows {
		title := win.Name
		if title == "" {
			title = "(untitled)"
		}
		fmt.Fprintf(w, "## W%d — %s\n\n", win.Index, title)
		fmt.Fprintln(w, "| Tab | Pane | Status | ID | PID | Duration | Command | TTY | Cwd |")
		fmt.Fprintln(w, "|-----|------|--------|----|-----|----------|---------|-----|-----|")
		for _, tab := range win.Tabs {
			for _, s := range tab.Sessions {
				fmt.Fprintf(w, "| %d | %d | %s | `%s` | %s | %s | `%s` | `%s` | %s |\n",
					tab.Index, s.Index,
					idleLabel(s.Idle),
					shortID(s.ID),
					derefInt(s.PID),
					mdCell(derefStr(s.Duration)),
					mdCell(derefStr(s.Command)),
					strings.TrimPrefix(s.TTY, "/dev/"),
					mdCell(derefStr(s.Cwd)),
				)
			}
		}
		fmt.Fprintln(w)
	}
	return nil
}

func mdCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	if s == "" {
		return "-"
	}
	return s
}

func renderSessionMarkdown(w io.Writer, snap *Snapshot, s *SnapshotSession) error {
	fmt.Fprintf(w, "# Session `%s`\n\n", s.ID)
	fmt.Fprintf(w, "- **status**: %s\n", idleLabel(s.Idle))
	fmt.Fprintf(w, "- **window**: %d · **tab**: %d · **pane**: %d\n", s.WindowIndex, s.TabIndex, s.Index)
	fmt.Fprintf(w, "- **name**: %s\n", s.Name)
	fmt.Fprintf(w, "- **tty**: `%s`\n", s.TTY)
	fmt.Fprintf(w, "- **pid**: %s · **shell_pid**: %s\n", derefInt(s.PID), derefInt(s.ShellPID))
	fmt.Fprintf(w, "- **command**: `%s`\n", derefStr(s.Command))
	fmt.Fprintf(w, "- **command_line**: `%s`\n", derefStr(s.CommandLine))
	fmt.Fprintf(w, "- **cwd**: %s\n", derefStr(s.Cwd))
	fmt.Fprintf(w, "- **start_time**: %s\n", derefStr(s.StartTime))
	fmt.Fprintf(w, "- **duration**: %s\n", derefStr(s.Duration))
	_ = snap
	return nil
}

func renderHTML(w io.Writer, snap *Snapshot) error {
	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<title>iTerm2 snapshot</title>
<style>
  body { font-family: ui-sans-serif, system-ui, sans-serif; margin: 1.5rem; color: #1a1a1a; }
  h1 { font-size: 1.25rem; }
  h2 { font-size: 1.05rem; margin-top: 1.5rem; }
  .meta { color: #666; font-size: 0.9rem; }
  table { border-collapse: collapse; width: 100%; font-size: 0.85rem; margin-top: 0.5rem; }
  th, td { border: 1px solid #ddd; padding: 0.35rem 0.5rem; text-align: left; vertical-align: top; }
  th { background: #f5f5f5; }
  .idle { color: #0a0; font-weight: 600; }
  .busy { color: #a60; font-weight: 600; }
  .unknown { color: #888; }
  code { font-size: 0.85em; }
</style>
</head>
<body>
`)
	fmt.Fprintf(w, "<h1>iTerm2 snapshot</h1>\n")
	fmt.Fprintf(w, `<p class="meta">captured %s · host %s · %d windows · %d tabs · %d sessions · idle %d · busy %d</p>`+"\n",
		html.EscapeString(snap.CapturedAt), html.EscapeString(snap.Host),
		snap.Summary.Windows, snap.Summary.Tabs, snap.Summary.Sessions,
		snap.Summary.Idle, snap.Summary.Busy)

	for _, win := range snap.Windows {
		title := win.Name
		if title == "" {
			title = "(untitled)"
		}
		fmt.Fprintf(w, "<h2>W%d — %s</h2>\n", win.Index, html.EscapeString(title))
		fmt.Fprintln(w, `<table>
<thead><tr><th>Tab</th><th>Pane</th><th>Status</th><th>ID</th><th>PID</th><th>Duration</th><th>Command</th><th>TTY</th><th>Cwd</th></tr></thead>
<tbody>`)
		for _, tab := range win.Tabs {
			for _, s := range tab.Sessions {
				lab := idleLabel(s.Idle)
				fmt.Fprintf(w, `<tr><td>%d</td><td>%d</td><td class="%s">%s</td><td><code>%s</code></td><td>%s</td><td>%s</td><td><code>%s</code></td><td><code>%s</code></td><td>%s</td></tr>`+"\n",
					tab.Index, s.Index,
					html.EscapeString(lab), html.EscapeString(lab),
					html.EscapeString(s.ID),
					html.EscapeString(derefInt(s.PID)),
					html.EscapeString(derefStr(s.Duration)),
					html.EscapeString(derefStr(s.Command)),
					html.EscapeString(strings.TrimPrefix(s.TTY, "/dev/")),
					html.EscapeString(derefStr(s.Cwd)),
				)
			}
		}
		fmt.Fprintln(w, `</tbody></table>`)
	}
	fmt.Fprintln(w, `</body></html>`)
	return nil
}

func renderSessionHTML(w io.Writer, snap *Snapshot, s *SnapshotSession) error {
	lab := idleLabel(s.Idle)
	fmt.Fprint(w, `<!DOCTYPE html><html lang="en"><head><meta charset="utf-8"/><title>Session status</title>
<style>body{font-family:system-ui,sans-serif;margin:1.5rem}.idle{color:#0a0}.busy{color:#a60}dt{font-weight:600;margin-top:0.4rem}dd{margin-left:1rem;color:#333}</style>
</head><body>`)
	fmt.Fprintf(w, `<h1>Session <code>%s</code></h1>`, html.EscapeString(s.ID))
	fmt.Fprintf(w, `<p class="%s">%s</p>`, html.EscapeString(lab), html.EscapeString(lab))
	fmt.Fprintln(w, "<dl>")
	fields := [][2]string{
		{"window", fmt.Sprintf("%d", s.WindowIndex)},
		{"tab", fmt.Sprintf("%d", s.TabIndex)},
		{"pane", fmt.Sprintf("%d", s.Index)},
		{"name", s.Name},
		{"tty", s.TTY},
		{"pid", derefInt(s.PID)},
		{"shell_pid", derefInt(s.ShellPID)},
		{"command", derefStr(s.Command)},
		{"command_line", derefStr(s.CommandLine)},
		{"cwd", derefStr(s.Cwd)},
		{"start_time", derefStr(s.StartTime)},
		{"duration", derefStr(s.Duration)},
	}
	for _, f := range fields {
		fmt.Fprintf(w, "<dt>%s</dt><dd>%s</dd>\n", html.EscapeString(f[0]), html.EscapeString(f[1]))
	}
	fmt.Fprintln(w, "</dl></body></html>")
	_ = snap
	return nil
}

// WriteWarning prints a yellow warning line to stderr when color is appropriate.
func WriteWarning(stderr io.Writer, msg string) {
	// Always prefix warning:; optional color if stderr is TTY
	line := msg
	if !strings.HasPrefix(msg, "warning:") {
		line = "warning: " + msg
	}
	if f, ok := stderr.(*os.File); ok && term.IsTerminal(int(f.Fd())) && os.Getenv("NO_COLOR") == "" {
		fmt.Fprintln(stderr, ansiYellow+line+ansiReset)
		return
	}
	fmt.Fprintln(stderr, line)
}

// WriteError prints a red Error line when appropriate.
func WriteError(stderr io.Writer, msg string) {
	line := msg
	if !strings.HasPrefix(msg, "Error:") && !strings.HasPrefix(msg, "error:") {
		line = "Error: " + msg
	}
	if f, ok := stderr.(*os.File); ok && term.IsTerminal(int(f.Fd())) && os.Getenv("NO_COLOR") == "" {
		fmt.Fprintln(stderr, ansiRed+line+ansiReset)
		return
	}
	fmt.Fprintln(stderr, line)
}
