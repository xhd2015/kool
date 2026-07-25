package iterm2

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/kool/pkgs/errs"
	"github.com/xhd2015/kool/pkgs/terminal"
	"golang.org/x/term"
)

const (
	sessionsSaveVersion = 1
	sessionsSaveSource  = "kool-iterm2-sessions-save"
)

// SaveDocument is the checkpoint written by sessions save / read by restore.
type SaveDocument struct {
	Version    int               `json:"version"`
	SavedAt    string            `json:"saved_at"`
	RestoredAt *string           `json:"restored_at"` // null until restore succeeds
	Host       string            `json:"host"`
	Source     string            `json:"source"`
	Summary    SaveSummary       `json:"summary"`
	Windows    []SaveWindow      `json:"windows"`
}

// SaveSummary counts critical tabs in the checkpoint.
type SaveSummary struct {
	Windows  int            `json:"windows"`
	Tabs     int            `json:"tabs"`
	Sessions int            `json:"sessions"`
	ByKind   map[string]int `json:"by_kind"`
}

// SaveWindow is one original iTerm window that had critical tabs.
type SaveWindow struct {
	SourceIndex int       `json:"source_index"`
	Name        string    `json:"name,omitempty"`
	Tabs        []SaveTab `json:"tabs"`
}

// SaveTab is one critical pane to restore as a tab (cd + resume_cmd).
type SaveTab struct {
	SourceTabIndex  int    `json:"source_tab_index,omitempty"`
	SourcePaneIndex int    `json:"source_pane_index,omitempty"`
	Name            string `json:"name,omitempty"`
	Cwd             string `json:"cwd"`
	Kind            string `json:"kind"` // grok | codex | mark
	SessionID       string `json:"session_id,omitempty"`
	Message         string `json:"message,omitempty"` // mark only
	Title           string `json:"title,omitempty"`
	ResumeCmd       string `json:"resume_cmd"`
	ItermSessionID  string `json:"iterm_session_id,omitempty"`
	SourceCmdLine   string `json:"source_command_line,omitempty"`
}

// DefaultSessionsSavePath is ~/.config/iterm2/sessions-save.json.
func DefaultSessionsSavePath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = "."
	}
	return filepath.Join(home, ".config", "iterm2", "sessions-save.json")
}

// Injectable hooks for tests.
var (
	sessionsNowFn          = time.Now
	sessionsHostnameFn     = os.Hostname
	sessionsIsStdinTTY     = terminal.IsStdinTerminal
	sessionsReadConfirm    = defaultReadConfirm
	sessionsRunRestoreAS   = defaultRunAppleScript
	sessionsSavePathForTest string // if set, overrides default path when --file omitted
)

func defaultReadConfirm(prompt string, stdout, stderr io.Writer) (string, error) {
	fmt.Fprint(stderr, prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// effectiveSavePath returns --file or default (test override aware).
func effectiveSavePath(fileFlag string) string {
	if fileFlag != "" {
		return fileFlag
	}
	if sessionsSavePathForTest != "" {
		return sessionsSavePathForTest
	}
	return DefaultSessionsSavePath()
}

// BuildSaveDocument filters a live snapshot to critical grok/codex/mark tabs.
// Skips panes with empty cwd (warning). Prefers agent over mark on the same pane.
func BuildSaveDocument(snap *Snapshot, now time.Time, host string) (*SaveDocument, []string) {
	var warnings []string
	if snap == nil {
		return emptySaveDoc(now, host), nil
	}
	if now.IsZero() {
		now = time.Now()
	}
	if host == "" {
		host, _ = os.Hostname()
	}

	byKind := map[string]int{}
	var windows []SaveWindow
	for _, win := range snap.Windows {
		var tabs []SaveTab
		for _, tab := range win.Tabs {
			for _, s := range tab.Sessions {
				st, warn := classifyCriticalTab(win, tab, s)
				if warn != "" {
					warnings = append(warnings, warn)
				}
				if st == nil {
					continue
				}
				tabs = append(tabs, *st)
				byKind[st.Kind]++
			}
		}
		if len(tabs) == 0 {
			continue
		}
		windows = append(windows, SaveWindow{
			SourceIndex: win.Index,
			Name:        win.Name,
			Tabs:        tabs,
		})
	}

	nTabs := 0
	for _, w := range windows {
		nTabs += len(w.Tabs)
	}
	doc := &SaveDocument{
		Version:    sessionsSaveVersion,
		SavedAt:    now.Format("2006-01-02T15:04:05-0700"),
		RestoredAt: nil,
		Host:       host,
		Source:     sessionsSaveSource,
		Summary: SaveSummary{
			Windows:  len(windows),
			Tabs:     nTabs,
			Sessions: nTabs,
			ByKind:   byKind,
		},
		Windows: windows,
	}
	return doc, warnings
}

func emptySaveDoc(now time.Time, host string) *SaveDocument {
	if now.IsZero() {
		now = time.Now()
	}
	return &SaveDocument{
		Version:    sessionsSaveVersion,
		SavedAt:    now.Format("2006-01-02T15:04:05-0700"),
		RestoredAt: nil,
		Host:       host,
		Source:     sessionsSaveSource,
		Summary:    SaveSummary{ByKind: map[string]int{}},
		Windows:    nil,
	}
}

func classifyCriticalTab(win SnapshotWindow, tab SnapshotTab, s SnapshotSession) (*SaveTab, string) {
	cwd := derefStr(s.Cwd)
	base := SaveTab{
		SourceTabIndex:  tab.Index,
		SourcePaneIndex: s.Index,
		Name:            firstNonEmpty(s.Name, tab.Name),
		Cwd:             cwd,
		ItermSessionID:  s.ID,
		SourceCmdLine:   derefStr(s.CommandLine),
	}

	// Prefer agent (grok/codex) over mark.
	if s.Agent != nil {
		kind := strings.ToLower(strings.TrimSpace(s.Agent.Kind))
		if (kind == "grok" || kind == "codex") && s.Agent.SessionID != "" {
			if cwd == "" {
				return nil, fmt.Sprintf("warning: skipped %s pane with empty cwd (iterm id %s)", kind, shortID(s.ID))
			}
			base.Kind = kind
			base.SessionID = s.Agent.SessionID
			base.Title = s.Agent.Title
			base.ResumeCmd = resumeCmdForAgent(kind, s.Agent.SessionID)
			return &base, ""
		}
	}

	if msg, ok := detectMarkSession(&s); ok {
		if cwd == "" {
			return nil, fmt.Sprintf("warning: skipped mark pane with empty cwd (iterm id %s)", shortID(s.ID))
		}
		base.Kind = "mark"
		base.Message = msg
		base.ResumeCmd = resumeCmdForMark(msg)
		return &base, ""
	}
	return nil, ""
}

func resumeCmdForAgent(kind, sessionID string) string {
	switch kind {
	case "codex":
		return "codex resume " + sessionID
	default:
		return "grok --resume " + sessionID
	}
}

func resumeCmdForMark(message string) string {
	if message == "" {
		return "mark"
	}
	return "mark " + shellSingleQuote(message)
}

// shellSingleQuote wraps s in single quotes safe for a POSIX shell.
func shellSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// detectMarkSession finds a live mark process on the pane.
func detectMarkSession(s *SnapshotSession) (message string, ok bool) {
	// Prefer process list (most accurate cmdline).
	for i := len(s.Processes) - 1; i >= 0; i-- {
		if isMarkCmdline(s.Processes[i].Command) {
			return markMessageFromCmdline(s.Processes[i].Command), true
		}
	}
	if s.Command != nil && isMarkBasename(*s.Command) {
		if s.CommandLine != nil && *s.CommandLine != "" {
			return markMessageFromCmdline(*s.CommandLine), true
		}
		return "", true
	}
	if s.CommandLine != nil && isMarkCmdline(*s.CommandLine) {
		return markMessageFromCmdline(*s.CommandLine), true
	}
	return "", false
}

func isMarkBasename(cmd string) bool {
	base := filepath.Base(strings.TrimSpace(cmd))
	return base == "mark"
}

func isMarkCmdline(cmdline string) bool {
	fields := strings.Fields(strings.TrimSpace(cmdline))
	if len(fields) == 0 {
		return false
	}
	return filepath.Base(fields[0]) == "mark"
}

func markMessageFromCmdline(cmdline string) string {
	fields := strings.Fields(strings.TrimSpace(cmdline))
	if len(fields) <= 1 {
		return ""
	}
	return strings.Join(fields[1:], " ")
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// IsConsumed returns true when restored_at is set to a non-empty string.
func (d *SaveDocument) IsConsumed() bool {
	return d != nil && d.RestoredAt != nil && strings.TrimSpace(*d.RestoredAt) != ""
}

// ReadSaveDocument loads and validates a checkpoint file.
func ReadSaveDocument(path string) (*SaveDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc SaveDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("invalid checkpoint JSON: %w", err)
	}
	if doc.Version != sessionsSaveVersion {
		return nil, fmt.Errorf("unsupported checkpoint version %d (want %d)", doc.Version, sessionsSaveVersion)
	}
	if doc.Summary.ByKind == nil {
		doc.Summary.ByKind = map[string]int{}
	}
	return &doc, nil
}

// WriteSaveDocument writes the checkpoint atomically (temp + rename).
func WriteSaveDocument(path string, doc *SaveDocument) error {
	if doc == nil {
		return fmt.Errorf("nil document")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// extractCriticalFromWindow classifies critical tabs in one live window.
// Returns a zero-tabs SaveWindow (and warnings) when nothing critical.
func extractCriticalFromWindow(win SnapshotWindow) (SaveWindow, []string) {
	var tabs []SaveTab
	var warnings []string
	for _, tab := range win.Tabs {
		for _, s := range tab.Sessions {
			st, warn := classifyCriticalTab(win, tab, s)
			if warn != "" {
				warnings = append(warnings, warn)
			}
			if st != nil {
				tabs = append(tabs, *st)
			}
		}
	}
	if len(tabs) == 0 {
		return SaveWindow{}, warnings
	}
	return SaveWindow{
		SourceIndex: win.Index,
		Name:        win.Name,
		Tabs:        tabs,
	}, warnings
}

// resolvePlanColor implements save/restore color policy (cli/color).
// --color and --no-color conflict; --color wins over NO_COLOR and non-TTY.
func resolvePlanColor(forceColor, forceNoColor bool) (bool, error) {
	if forceColor && forceNoColor {
		return false, fmt.Errorf("--color and --no-color cannot be specified together")
	}
	if forceColor {
		return true, nil
	}
	if forceNoColor {
		return false, nil
	}
	if os.Getenv("NO_COLOR") != "" {
		return false, nil
	}
	return term.IsTerminal(int(os.Stdout.Fd())), nil
}

func paintKind(color bool, kind string) string {
	switch strings.ToLower(kind) {
	case "grok", "codex":
		return paint(color, ansiGreen, kind)
	case "mark":
		return paint(color, ansiYellow, kind)
	default:
		return kind
	}
}

// formatSaveWindowBlock writes one critical window plan block (streaming dry-run).
func formatSaveWindowBlock(w io.Writer, win SaveWindow, color bool) {
	label := win.Name
	if label == "" {
		label = "(untitled)"
	}
	wLabel := paint(color, ansiBold, fmt.Sprintf("W%d", win.SourceIndex))
	fmt.Fprintf(w, "\n  %s  %q\n", wLabel, label)
	for i, tab := range win.Tabs {
		id := tab.SessionID
		if tab.Kind == "mark" {
			id = tab.Message
			if id == "" {
				id = "(no message)"
			}
		}
		kind := paintKind(color, tab.Kind)
		// Pad plain kind for monochrome column alignment; colored codes skip pad.
		kindCol := kind
		if !color {
			kindCol = fmt.Sprintf("%-5s", tab.Kind)
		}
		cwd := paint(color, ansiGray, tab.Cwd)
		fmt.Fprintf(w, "      tab%d  %s  %s  %s\n", i+1, kindCol, id, cwd)
		fmt.Fprintf(w, "            %s\n", paint(color, ansiGray, tab.ResumeCmd))
	}
}

// formatSaveFooter writes the Would save / Saved summary, path, and dry-run note.
func formatSaveFooter(w io.Writer, doc *SaveDocument, path string, dryRun bool, color bool) {
	kindParts := make([]string, 0, len(doc.Summary.ByKind))
	for _, k := range []string{"grok", "codex", "mark"} {
		if n := doc.Summary.ByKind[k]; n > 0 {
			kindParts = append(kindParts, fmt.Sprintf("%d %s", n, paintKind(color, k)))
		}
	}
	for k, n := range doc.Summary.ByKind {
		if k == "grok" || k == "codex" || k == "mark" {
			continue
		}
		if n > 0 {
			kindParts = append(kindParts, fmt.Sprintf("%d %s", n, paintKind(color, k)))
		}
	}
	kindStr := strings.Join(kindParts, ", ")
	if kindStr == "" {
		kindStr = "none"
	}

	verb := "Saved"
	if dryRun {
		verb = "Would save"
	}
	fmt.Fprintf(w, "%s %d critical sessions (%s) in %d windows\n",
		paint(color, ansiGreen, verb), doc.Summary.Sessions, kindStr, doc.Summary.Windows)
	fmt.Fprintf(w, "  → %s\n", paint(color, ansiGray, path))
	if dryRun {
		fmt.Fprintln(w, paint(color, ansiGray, "(dry-run: not written)"))
	}
}

// formatSavePlan writes a human plan of the document to w.
// Dry-run: window blocks then footer (matches progressive stream order).
// Live save: footer only.
func formatSavePlan(w io.Writer, doc *SaveDocument, path string, dryRun bool, color bool) {
	if dryRun {
		for _, win := range doc.Windows {
			formatSaveWindowBlock(w, win, color)
		}
	}
	formatSaveFooter(w, doc, path, dryRun, color)
}

// runSessionsSave implements `sessions save`.
func runSessionsSave(args []string, stdout, stderr io.Writer) error {
	var dryRun bool
	var fileFlag string
	var forceColor, forceNoColor bool
	remain, err := parseSaveRestoreFlags(args, &dryRun, &fileFlag, &forceColor, &forceNoColor, sessionsSaveHelp)
	if err != nil {
		if err == errHelpRequested {
			fmt.Fprint(stdout, strings.TrimSpace(sessionsSaveHelp)+"\n")
			return nil
		}
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) > 0 {
		fmt.Fprintf(stderr, "Error: sessions save: unexpected arguments: %s\n", strings.Join(remain, " "))
		return errs.NewSilenceExitCode(1)
	}

	color, err := resolvePlanColor(forceColor, forceNoColor)
	if err != nil {
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}

	path := effectiveSavePath(fileFlag)

	// Dry-run: progressive capture — emit each critical window block as ready.
	if dryRun {
		return runSessionsSaveDryRun(stdout, stderr, path, color)
	}

	snap, warnings, err := CaptureSnapshotWith(CaptureOpts{NoEnrich: false})
	if err != nil {
		WriteError(stderr, strings.TrimPrefix(err.Error(), "Error: "))
		return errs.NewSilenceExitCode(1)
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

	if doc.Summary.Sessions == 0 {
		fmt.Fprintln(stdout, "0 critical sessions (nothing to save)")
		return nil
	}

	// Overwrite policy when file exists.
	if existing, rerr := ReadSaveDocument(path); rerr == nil {
		if !existing.IsConsumed() {
			if !sessionsIsStdinTTY() {
				WriteError(stderr, fmt.Sprintf(
					"sessions save: checkpoint exists and is not restored yet; confirm overwrite on a TTY (or remove the file)\n  path: %s\n  saved_at: %s",
					path, existing.SavedAt))
				return errs.NewSilenceExitCode(1)
			}
			prompt := fmt.Sprintf(
				"Overwrite existing save (saved_at=%s, %d sessions, not restored yet)? [Y/n] ",
				existing.SavedAt, existing.Summary.Sessions)
			ans, cerr := sessionsReadConfirm(prompt, stdout, stderr)
			if cerr != nil {
				WriteError(stderr, cerr.Error())
				return errs.NewSilenceExitCode(1)
			}
			// Default Y on empty; decline on n/N/no.
			low := strings.ToLower(strings.TrimSpace(ans))
			if low == "n" || low == "no" {
				fmt.Fprintln(stderr, "sessions save: overwrite declined")
				return nil
			}
		}
		// Already restored: overwrite without prompt.
	} else if !os.IsNotExist(rerr) {
		// Unreadable existing: treat as overwrite target if we can write.
		// If not JSON, still attempt write after warning.
		if !os.IsNotExist(rerr) && !strings.Contains(rerr.Error(), "no such file") {
			// path exists but bad JSON — prompt like pending if TTY else error
			if _, statErr := os.Stat(path); statErr == nil {
				if !sessionsIsStdinTTY() {
					WriteError(stderr, fmt.Sprintf(
						"sessions save: checkpoint exists (unreadable) and cannot confirm overwrite without a TTY\n  path: %s\n  detail: %v",
						path, rerr))
					return errs.NewSilenceExitCode(1)
				}
			}
		}
	}

	// Fresh save: clear restored_at.
	doc.RestoredAt = nil
	if err := WriteSaveDocument(path, doc); err != nil {
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	formatSavePlan(stdout, doc, path, false, color)
	return nil
}

// runSessionsSaveDryRun captures windows progressively, writing each critical
// window block as soon as that window is ready, then the Would save footer.
func runSessionsSaveDryRun(stdout, stderr io.Writer, path string, color bool) error {
	now := sessionsNowFn()
	host, _ := sessionsHostnameFn()

	var windows []SaveWindow
	byKind := map[string]int{}
	var filterWarns []string

	c := activeCollector()
	_, warnings, err := c.capture(func(win SnapshotWindow) error {
		sw, warns := extractCriticalFromWindow(win)
		filterWarns = append(filterWarns, warns...)
		if len(sw.Tabs) == 0 {
			return nil
		}
		for _, tab := range sw.Tabs {
			byKind[tab.Kind]++
		}
		windows = append(windows, sw)
		formatSaveWindowBlock(stdout, sw, color)
		return nil
	}, CaptureOpts{NoEnrich: false})
	if err != nil {
		WriteError(stderr, strings.TrimPrefix(err.Error(), "Error: "))
		return errs.NewSilenceExitCode(1)
	}
	for _, w := range warnings {
		WriteWarning(stderr, strings.TrimPrefix(w, "warning: "))
	}
	for _, w := range filterWarns {
		WriteWarning(stderr, strings.TrimPrefix(w, "warning: "))
	}

	nTabs := 0
	for _, w := range windows {
		nTabs += len(w.Tabs)
	}
	if nTabs == 0 {
		fmt.Fprintln(stdout, "0 critical sessions (nothing to save)")
		return nil
	}
	if byKind == nil {
		byKind = map[string]int{}
	}
	doc := &SaveDocument{
		Version:    sessionsSaveVersion,
		SavedAt:    now.Format("2006-01-02T15:04:05-0700"),
		RestoredAt: nil,
		Host:       host,
		Source:     sessionsSaveSource,
		Summary: SaveSummary{
			Windows:  len(windows),
			Tabs:     nTabs,
			Sessions: nTabs,
			ByKind:   byKind,
		},
		Windows: windows,
	}
	formatSaveFooter(stdout, doc, path, true, color)
	return nil
}

// runSessionsRestore implements `sessions restore`.
func runSessionsRestore(args []string, stdout, stderr io.Writer) error {
	var dryRun bool
	var fileFlag string
	var forceColor, forceNoColor bool
	remain, err := parseSaveRestoreFlags(args, &dryRun, &fileFlag, &forceColor, &forceNoColor, sessionsRestoreHelp)
	if err != nil {
		if err == errHelpRequested {
			fmt.Fprint(stdout, strings.TrimSpace(sessionsRestoreHelp)+"\n")
			return nil
		}
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) > 0 {
		fmt.Fprintf(stderr, "Error: sessions restore: unexpected arguments: %s\n", strings.Join(remain, " "))
		return errs.NewSilenceExitCode(1)
	}

	color, err := resolvePlanColor(forceColor, forceNoColor)
	if err != nil {
		WriteError(stderr, err.Error())
		return errs.NewSilenceExitCode(1)
	}

	path := effectiveSavePath(fileFlag)
	doc, err := ReadSaveDocument(path)
	if err != nil {
		if os.IsNotExist(err) {
			WriteError(stderr, fmt.Sprintf(
				"sessions restore: checkpoint not found: %s\n  save first: kool iterm2 sessions save", path))
			return errs.NewSilenceExitCode(1)
		}
		WriteError(stderr, fmt.Sprintf("sessions restore: %v", err))
		return errs.NewSilenceExitCode(1)
	}

	if doc.IsConsumed() {
		WriteError(stderr, fmt.Sprintf(
			"sessions restore: checkpoint already consumed (restored_at=%s)\n  path: %s\n  re-save with: kool iterm2 sessions save",
			*doc.RestoredAt, path))
		return errs.NewSilenceExitCode(1)
	}

	if doc.Summary.Sessions == 0 || len(doc.Windows) == 0 {
		fmt.Fprintln(stdout, "0 critical sessions in checkpoint (nothing to restore)")
		return nil
	}

	host, _ := sessionsHostnameFn()
	if doc.Host != "" && host != "" && doc.Host != host {
		WriteWarning(stderr, fmt.Sprintf("host in file is %q, this machine is %q", doc.Host, host))
	}

	if dryRun {
		formatRestorePlan(stdout, doc, path, true, color)
		return nil
	}

	script := BuildSessionsRestoreScript(doc)
	if _, err := sessionsRunRestoreAS(script); err != nil {
		WriteError(stderr, fmt.Sprintf("sessions restore: AppleScript failed: %v", err))
		return errs.NewSilenceExitCode(1)
	}

	now := sessionsNowFn()
	ts := now.Format("2006-01-02T15:04:05-0700")
	doc.RestoredAt = &ts
	if err := WriteSaveDocument(path, doc); err != nil {
		WriteError(stderr, fmt.Sprintf("sessions restore: windows created but failed to mark restored_at: %v", err))
		return errs.NewSilenceExitCode(1)
	}
	formatRestorePlan(stdout, doc, path, false, color)
	return nil
}

func formatRestorePlan(w io.Writer, doc *SaveDocument, path string, dryRun bool, color bool) {
	if dryRun {
		fmt.Fprintf(w, "%s %d windows / %d tabs from %s\n",
			paint(color, ansiGreen, "Would restore"),
			doc.Summary.Windows, doc.Summary.Tabs,
			paint(color, ansiGray, path))
		fmt.Fprintf(w, "  %s\n", paint(color, ansiGray, fmt.Sprintf("(saved_at %s, not yet restored)", doc.SavedAt)))
		for _, win := range doc.Windows {
			fmt.Fprintf(w, "\n  %s\n", paint(color, ansiBold, "new window"))
			for _, tab := range win.Tabs {
				fmt.Fprintf(w, "    tab  %s\n", paint(color, ansiGray, "cd "+shellSingleQuote(tab.Cwd)))
				fmt.Fprintf(w, "         %s\n", paint(color, ansiGray, tab.ResumeCmd))
			}
		}
		fmt.Fprintln(w, paint(color, ansiGray, "(dry-run: not applied)"))
		return
	}
	fmt.Fprintf(w, "%s %d windows / %d tabs\n",
		paint(color, ansiGreen, "Restored"), doc.Summary.Windows, doc.Summary.Tabs)
	if doc.RestoredAt != nil {
		fmt.Fprintf(w, "  %s\n", paint(color, ansiGray, fmt.Sprintf("marked restored_at=%s", *doc.RestoredAt)))
	}
	fmt.Fprintf(w, "  → %s\n", paint(color, ansiGray, path))
}

// errHelpRequested is returned when -h was parsed.
var errHelpRequested = fmt.Errorf("help")

func parseSaveRestoreFlags(args []string, dryRun *bool, fileFlag *string, forceColor, forceNoColor *bool, helpText string) ([]string, error) {
	// Manual parse to keep less-flags optional and help clean.
	var remain []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help" || a == "help":
			return nil, errHelpRequested
		case a == "--dry-run":
			*dryRun = true
		case a == "--color":
			*forceColor = true
		case a == "--no-color":
			*forceNoColor = true
		case a == "--file" || a == "-f":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--file requires a path")
			}
			i++
			*fileFlag = args[i]
		case strings.HasPrefix(a, "--file="):
			*fileFlag = strings.TrimPrefix(a, "--file=")
		default:
			if strings.HasPrefix(a, "-") {
				return nil, fmt.Errorf("unknown flag %s", a)
			}
			remain = append(remain, a)
		}
	}
	_ = helpText
	return remain, nil
}
