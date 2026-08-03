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
	SourceIndex   int       `json:"source_index"`
	Name          string    `json:"name,omitempty"`
	Space         int       `json:"space"`                         // 0-based Desktop; always emitted when not ignore
	ItermWindowID int64     `json:"iterm_window_id,omitempty"`     // info only at save; restore never uses it
	Tabs          []SaveTab `json:"tabs"`
	// noSpaceRecord omits space + iterm_window_id on marshal (--ignore-macos-space).
	noSpaceRecord bool `json:"-"`
}

// MarshalJSON always emits "space" (including 0) unless noSpaceRecord.
func (w SaveWindow) MarshalJSON() ([]byte, error) {
	type out struct {
		SourceIndex   int       `json:"source_index"`
		Name          string    `json:"name,omitempty"`
		Space         *int      `json:"space,omitempty"`
		ItermWindowID int64     `json:"iterm_window_id,omitempty"`
		Tabs          []SaveTab `json:"tabs"`
	}
	o := out{
		SourceIndex: w.SourceIndex,
		Name:        w.Name,
		Tabs:        w.Tabs,
	}
	if !w.noSpaceRecord {
		s := w.Space
		o.Space = &s
		o.ItermWindowID = w.ItermWindowID
	}
	return json.Marshal(o)
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
	sessionsNowFn           = time.Now
	sessionsHostnameFn      = os.Hostname
	sessionsIsStdinTTY      = terminal.IsStdinTerminal
	sessionsReadConfirm     = defaultReadConfirm
	sessionsRunRestoreAS    = defaultRunAppleScript
	sessionsSavePathForTest string // if set, overrides default path when --file omitted
)

// SetSessionsRunRestoreASForTest installs the restore AppleScript runner for L2/unit tests.
// Pass nil to restore the production default. Prefer t.Cleanup to restore.
func SetSessionsRunRestoreASForTest(fn func(script string) (string, error)) {
	if fn == nil {
		sessionsRunRestoreAS = defaultRunAppleScript
		return
	}
	sessionsRunRestoreAS = fn
}

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

// applySpaceToSaveDocument fills Space / ItermWindowID from the live snapshot.
// When ignore is true, marks windows to omit both fields (no resolve).
func applySpaceToSaveDocument(doc *SaveDocument, snap *Snapshot, ignore bool) []string {
	if doc == nil {
		return nil
	}
	byIndex := map[int]SnapshotWindow{}
	if snap != nil {
		for _, w := range snap.Windows {
			byIndex[w.Index] = w
		}
	}
	var warnings []string
	for i := range doc.Windows {
		sw := &doc.Windows[i]
		win, ok := byIndex[sw.SourceIndex]
		if !ok {
			win = SnapshotWindow{Index: sw.SourceIndex}
		}
		if warn := attachSpaceFields(sw, win, ignore); warn != "" {
			warnings = append(warnings, "warning: "+warn)
		}
	}
	return warnings
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
// ignoreSpace skips Space resolve and omits space fields on marshal.
func extractCriticalFromWindow(win SnapshotWindow, ignoreSpace bool) (SaveWindow, []string) {
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
	sw := SaveWindow{
		SourceIndex: win.Index,
		Name:        win.Name,
		Tabs:        tabs,
	}
	if warn := attachSpaceFields(&sw, win, ignoreSpace); warn != "" {
		warnings = append(warnings, "warning: "+warn)
	}
	return sw, warnings
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
	if !win.noSpaceRecord {
		// space N (Desktop N+1); never show iterm_window_id in plan text.
		meta := formatSpaceDesktopLabel(win.Space)
		fmt.Fprintf(w, "    %s\n", paint(color, ansiGray, meta))
	}
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
	var ignoreMacOSSpace bool
	remain, err := parseSaveRestoreFlags(args, &dryRun, &fileFlag, &forceColor, &forceNoColor, &ignoreMacOSSpace, sessionsSaveHelp)
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
		return runSessionsSaveDryRun(stdout, stderr, path, color, ignoreMacOSSpace)
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
	for _, w := range applySpaceToSaveDocument(doc, snap, ignoreMacOSSpace) {
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
func runSessionsSaveDryRun(stdout, stderr io.Writer, path string, color bool, ignoreMacOSSpace bool) error {
	now := sessionsNowFn()
	host, _ := sessionsHostnameFn()

	var windows []SaveWindow
	byKind := map[string]int{}
	var filterWarns []string

	c := activeCollector()
	_, warnings, err := c.capture(func(win SnapshotWindow) error {
		sw, warns := extractCriticalFromWindow(win, ignoreMacOSSpace)
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

// liveCriticalHit is one live critical pane used for already-running match.
type liveCriticalHit struct {
	Kind  string // grok | codex | mark
	ID    string // session_id or mark message
	Name  string
	PID   int
	Space int // 0-based Desktop index (soft: 0 when unresolved)
}

// criticalMatchKey returns the already-running lookup key for a save tab.
// Agent: "kind:session_id"; mark: "mark:message". Empty when unmatchable.
func criticalMatchKey(tab SaveTab) string {
	kind := strings.ToLower(strings.TrimSpace(tab.Kind))
	switch kind {
	case "grok", "codex":
		if tab.SessionID == "" {
			return ""
		}
		return kind + ":" + tab.SessionID
	case "mark":
		return "mark:" + tab.Message
	default:
		return ""
	}
}

// criticalIdentity returns the identity string shown in already-running warnings.
func criticalIdentity(tab SaveTab) string {
	kind := strings.ToLower(strings.TrimSpace(tab.Kind))
	if kind == "mark" {
		return tab.Message
	}
	return tab.SessionID
}

// pidForLiveCritical prefers agent tree / mark process, then pane pid / shell.
func pidForLiveCritical(s *SnapshotSession, st *SaveTab) int {
	if s == nil {
		return 0
	}
	if s.Agent != nil && len(s.Agent.Tree) > 0 {
		kind := strings.ToLower(st.Kind)
		for i := len(s.Agent.Tree) - 1; i >= 0; i-- {
			n := s.Agent.Tree[i]
			role := strings.ToLower(n.Role)
			if role == kind || strings.Contains(strings.ToLower(n.Cmd), kind) {
				return n.PID
			}
		}
		return s.Agent.Tree[len(s.Agent.Tree)-1].PID
	}
	if st.Kind == "mark" {
		for i := len(s.Processes) - 1; i >= 0; i-- {
			if isMarkCmdline(s.Processes[i].Command) {
				return s.Processes[i].PID
			}
		}
	}
	if s.PID != nil && *s.PID > 0 {
		return *s.PID
	}
	if s.ShellPID != nil && *s.ShellPID > 0 {
		return *s.ShellPID
	}
	return 0
}

// indexLiveCritical walks a live snapshot and indexes first-hit critical panes
// by kind+session_id / mark message (D3, D5, D8).
func indexLiveCritical(snap *Snapshot) map[string]liveCriticalHit {
	idx := make(map[string]liveCriticalHit)
	if snap == nil {
		return idx
	}
	for _, win := range snap.Windows {
		spaceIdx, _, _ := resolveSpaceForWindow(win) // soft on fail → 0
		for _, tab := range win.Tabs {
			for i := range tab.Sessions {
				s := &tab.Sessions[i]
				st, _ := classifyCriticalTab(win, tab, *s)
				if st == nil {
					continue
				}
				key := criticalMatchKey(*st)
				if key == "" {
					continue
				}
				if _, exists := idx[key]; exists {
					continue // first live hit only
				}
				name := firstNonEmpty(st.Name, firstNonEmpty(s.Name, tab.Name))
				idx[key] = liveCriticalHit{
					Kind:  st.Kind,
					ID:    criticalIdentity(*st),
					Name:  name,
					PID:   pidForLiveCritical(s, st),
					Space: spaceIdx,
				}
			}
		}
	}
	return idx
}

// formatAlreadyRunningWarning matches D4:
//
//	tab "<name>" (kind id) is already running at space N (Desktop N+1), pid P
func formatAlreadyRunningWarning(tab SaveTab, hit liveCriticalHit) string {
	name := tab.Name
	if name == "" {
		name = hit.Name
	}
	kind := strings.ToLower(strings.TrimSpace(tab.Kind))
	if kind == "" {
		kind = strings.ToLower(hit.Kind)
	}
	id := criticalIdentity(tab)
	if id == "" {
		id = hit.ID
	}
	space := hit.Space
	if space < 0 {
		space = 0
	}
	return fmt.Sprintf("tab %q (%s %s) is already running at %s, pid %d",
		name, kind, id, formatSpaceDesktopLabel(space), hit.PID)
}

// matchCheckpointSkips marks checkpoint tabs that are already live.
// Emits already-running warnings on hit. Returns tabSkipped[wi][ti] and skip count.
func matchCheckpointSkips(doc *SaveDocument, live map[string]liveCriticalHit, stderr io.Writer) (tabSkipped [][]bool, skipped int) {
	if doc == nil {
		return nil, 0
	}
	tabSkipped = make([][]bool, len(doc.Windows))
	for wi, win := range doc.Windows {
		tabSkipped[wi] = make([]bool, len(win.Tabs))
		for ti, tab := range win.Tabs {
			key := criticalMatchKey(tab)
			if key == "" {
				continue
			}
			hit, ok := live[key]
			if !ok {
				continue
			}
			tabSkipped[wi][ti] = true
			skipped++
			WriteWarning(stderr, formatAlreadyRunningWarning(tab, hit))
		}
	}
	return tabSkipped, skipped
}

// countRemainingWouldCreate returns would-create window/tab counts (skipped excluded).
func countRemainingWouldCreate(doc *SaveDocument, tabSkipped [][]bool) (windows, tabs int) {
	if doc == nil {
		return 0, 0
	}
	for wi, win := range doc.Windows {
		n := 0
		for ti := range win.Tabs {
			if tabSkipped != nil && wi < len(tabSkipped) && ti < len(tabSkipped[wi]) && tabSkipped[wi][ti] {
				continue
			}
			n++
		}
		if n > 0 {
			windows++
			tabs += n
		}
	}
	return windows, tabs
}

// filterSaveDocRemaining returns a SaveDocument with only non-skipped tabs;
// windows that become empty are omitted (E3).
func filterSaveDocRemaining(doc *SaveDocument, tabSkipped [][]bool) *SaveDocument {
	if doc == nil {
		return nil
	}
	out := &SaveDocument{
		Version:    doc.Version,
		SavedAt:    doc.SavedAt,
		RestoredAt: doc.RestoredAt,
		Host:       doc.Host,
		Source:     doc.Source,
		Summary:    doc.Summary,
	}
	for wi, win := range doc.Windows {
		var tabs []SaveTab
		for ti, tab := range win.Tabs {
			if tabSkipped != nil && wi < len(tabSkipped) && ti < len(tabSkipped[wi]) && tabSkipped[wi][ti] {
				continue
			}
			tabs = append(tabs, tab)
		}
		if len(tabs) == 0 {
			continue
		}
		w := win
		w.Tabs = tabs
		out.Windows = append(out.Windows, w)
	}
	return out
}

// runSessionsRestore implements `sessions restore`.
func runSessionsRestore(args []string, stdout, stderr io.Writer) error {
	var dryRun bool
	var fileFlag string
	var forceColor, forceNoColor bool
	var ignoreMacOSSpace bool
	remain, err := parseSaveRestoreFlags(args, &dryRun, &fileFlag, &forceColor, &forceNoColor, &ignoreMacOSSpace, sessionsRestoreHelp)
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

	// Already-running scan (D2 dry-run + live; D6 soft capture fail → 0 hits).
	live := map[string]liveCriticalHit{}
	snap, snapWarns, capErr := CaptureSnapshotWith(CaptureOpts{NoEnrich: false})
	if capErr != nil {
		msg := strings.TrimPrefix(capErr.Error(), "Error: ")
		msg = strings.TrimSpace(msg)
		WriteWarning(stderr, fmt.Sprintf("could not scan live sessions for already-running check: %s", msg))
	} else {
		for _, w := range snapWarns {
			WriteWarning(stderr, strings.TrimPrefix(w, "warning: "))
		}
		live = indexLiveCritical(snap)
	}
	tabSkipped, skippedCount := matchCheckpointSkips(doc, live, stderr)
	remainWindows, remainTabs := countRemainingWouldCreate(doc, tabSkipped)

	// Clamp warnings even on dry-run (no Switch/Create).
	if !ignoreMacOSSpace {
		for _, win := range doc.Windows {
			if _, warn := clampSpaceIndex(win.Space); warn != "" {
				WriteWarning(stderr, warn)
			}
		}
	}

	if dryRun {
		formatRestorePlan(stdout, doc, path, true, color, ignoreMacOSSpace, tabSkipped, skippedCount, remainWindows, remainTabs)
		return nil
	}

	// Live: all remaining 0 → still stamp restored_at; no AS create (E1).
	if remainTabs == 0 {
		now := sessionsNowFn()
		ts := now.Format("2006-01-02T15:04:05-0700")
		doc.RestoredAt = &ts
		if err := WriteSaveDocument(path, doc); err != nil {
			WriteError(stderr, fmt.Sprintf("sessions restore: failed to mark restored_at: %v", err))
			return errs.NewSilenceExitCode(1)
		}
		formatRestorePlan(stdout, doc, path, false, color, ignoreMacOSSpace, tabSkipped, skippedCount, remainWindows, remainTabs)
		return nil
	}

	// Live restore: only non-skipped tabs; omit all-skipped windows (E3).
	filtered := filterSaveDocRemaining(doc, tabSkipped)
	for _, win := range filtered.Windows {
		if !ignoreMacOSSpace {
			placeWarns, perr := ensureSpacePlacement(win.Space)
			for _, w := range placeWarns {
				WriteWarning(stderr, w)
			}
			if perr != nil {
				WriteError(stderr, fmt.Sprintf("sessions restore: Space placement failed: %v", perr))
				return errs.NewSilenceExitCode(1)
			}
		}
		// One window at a time so each lands on the frontmost Space.
		single := &SaveDocument{Windows: []SaveWindow{win}}
		script := BuildSessionsRestoreScript(single)
		asOut, aerr := sessionsRunRestoreAS(script)
		if aerr != nil {
			WriteError(stderr, fmt.Sprintf("sessions restore: AppleScript failed: %v", aerr))
			return errs.NewSilenceExitCode(1)
		}
		// Title failures are best-effort: AS continues restore and returns one warning line per title.
		for _, line := range strings.Split(asOut, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			WriteWarning(stderr, line)
		}
	}

	now := sessionsNowFn()
	ts := now.Format("2006-01-02T15:04:05-0700")
	doc.RestoredAt = &ts
	if err := WriteSaveDocument(path, doc); err != nil {
		WriteError(stderr, fmt.Sprintf("sessions restore: windows created but failed to mark restored_at: %v", err))
		return errs.NewSilenceExitCode(1)
	}
	formatRestorePlan(stdout, doc, path, false, color, ignoreMacOSSpace, tabSkipped, skippedCount, remainWindows, remainTabs)
	return nil
}

// formatRestorePlan writes the restore dry-run plan or live summary.
// Header/summary counts are would-create / actually restored (skipped excluded).
// Dry-run lists every checkpoint tab with skip / would-restore action markers (E2).
// When skippedCount > 0, a skip meta clause is included (E4).
func formatRestorePlan(w io.Writer, doc *SaveDocument, path string, dryRun bool, color bool, ignoreMacOSSpace bool, tabSkipped [][]bool, skippedCount, remainWindows, remainTabs int) {
	if dryRun {
		fmt.Fprintf(w, "%s %d windows / %d tabs from %s\n",
			paint(color, ansiGreen, "Would restore"),
			remainWindows, remainTabs,
			paint(color, ansiGray, path))
		fmt.Fprintf(w, "  %s\n", paint(color, ansiGray, fmt.Sprintf("(saved_at %s, not yet restored)", doc.SavedAt)))
		if skippedCount > 0 {
			fmt.Fprintf(w, "  %s\n", paint(color, ansiGray, fmt.Sprintf("(%d skipped already running)", skippedCount)))
		}
		for wi, win := range doc.Windows {
			fmt.Fprintf(w, "\n  %s\n", paint(color, ansiBold, "new window"))
			if !ignoreMacOSSpace {
				s, _ := clampSpaceIndex(win.Space)
				meta := formatSpaceDesktopLabel(s)
				fmt.Fprintf(w, "    %s\n", paint(color, ansiGray, meta))
			}
			for ti, tab := range win.Tabs {
				skipped := tabSkipped != nil && wi < len(tabSkipped) && ti < len(tabSkipped[wi]) && tabSkipped[wi][ti]
				if skipped {
					fmt.Fprintf(w, "    tab  %s\n", paint(color, ansiGray, "skip (already running)"))
					if tab.ResumeCmd != "" {
						fmt.Fprintf(w, "         %s\n", paint(color, ansiGray, tab.ResumeCmd))
					}
					continue
				}
				fmt.Fprintf(w, "    tab  %s\n", paint(color, ansiGray, "would restore"))
				fmt.Fprintf(w, "         %s\n", paint(color, ansiGray, "cd "+shellSingleQuote(tab.Cwd)))
				fmt.Fprintf(w, "         %s\n", paint(color, ansiGray, tab.ResumeCmd))
			}
		}
		fmt.Fprintln(w, paint(color, ansiGray, "(dry-run: not applied)"))
		return
	}

	header := fmt.Sprintf("%s %d windows / %d tabs",
		paint(color, ansiGreen, "Restored"), remainWindows, remainTabs)
	if skippedCount > 0 {
		header += fmt.Sprintf(" (%d skipped already running)", skippedCount)
	}
	fmt.Fprintln(w, header)
	if doc.RestoredAt != nil {
		fmt.Fprintf(w, "  %s\n", paint(color, ansiGray, fmt.Sprintf("marked restored_at=%s", *doc.RestoredAt)))
	}
	fmt.Fprintf(w, "  → %s\n", paint(color, ansiGray, path))
}

// errHelpRequested is returned when -h was parsed.
var errHelpRequested = fmt.Errorf("help")

func parseSaveRestoreFlags(args []string, dryRun *bool, fileFlag *string, forceColor, forceNoColor *bool, ignoreMacOSSpace *bool, helpText string) ([]string, error) {
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
		case a == "--ignore-macos-space":
			if ignoreMacOSSpace != nil {
				*ignoreMacOSSpace = true
			}
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
