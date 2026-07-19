package iterm2

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	lib "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
	"github.com/xhd2015/kool/pkgs/errs"
	"github.com/xhd2015/kool/pkgs/terminal"
	lessflags "github.com/xhd2015/less-flags"
)

const tabSetDirEnv = "KOOL_ITERM2_TAB_SET_DIR"

const tabSetHelp = `iterm2 tab-set — manage named multi-tab iTerm2 layouts

Usage:
  kool iterm2 tab-set list
  kool iterm2 tab-set show <name>
  kool iterm2 tab-set run <name> [run-options]
  kool iterm2 tab-set status <name>
  kool iterm2 tab-set stop <name>
  kool iterm2 tab-set -h|--help

Config:
  JSON files in ~/.config/iterm2/tab-set/<name>.json
  Override directory with env KOOL_ITERM2_TAB_SET_DIR

Commands:
  list                     list configured tab sets
  show <name>              print window name and tabs for a set
  run <name>               open or resync the tab set in iTerm2
  status <name>            report idle/running/missing tabs
  stop <name>              close windows/tabs for the set

Run options:
  --dry-run                print the plan without calling iTerm2 (or without writing on --save)
  -n, --new-window         always create a new window (not with --save)
  --no-new-window          sync into the frontmost window only (not with --save)
  --tab <spec>             ad-hoc tab (repeatable); do not read config JSON
  --save                   write ad-hoc tabs to <name>.json (never runs iTerm; requires --tab)
  --force                  with --save, skip y/N overwrite confirm
  --window-name <name>     optional window name for ad-hoc / save
  -h, --help               show this help

Ad-hoc --tab spec:
  [id=…,name=…,cwd=…] command
  Props optional; spaces around [ ] keys = and , are allowed.
  Default id: tab-1..tab-N (1-based --tab order); name defaults to id.
  --save requires ≥1 --tab. --force only valid with --save.
  To run after save: kool iterm2 tab-set run <name>

Examples:
  kool iterm2 tab-set list
  kool iterm2 tab-set show bots
  kool iterm2 tab-set run bots --dry-run
  kool iterm2 tab-set run bots -n
  kool iterm2 tab-set run scratch --tab "echo a" --tab "echo b" --dry-run
  kool iterm2 tab-set run bots --tab "[id=a] echo a" --save --force
  kool iterm2 tab-set status bots
  kool iterm2 tab-set stop bots
`

// tabSetFile is the on-disk version-1 schema.
type tabSetFile struct {
	Version    int          `json:"version"`
	WindowName string       `json:"window_name"`
	Tabs       []tabSetTab  `json:"tabs"`
}

type tabSetTab struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Command string `json:"command"`
	Cwd     string `json:"cwd"`
}

// loadedTabSet is a validated config ready for orchestration.
type loadedTabSet struct {
	Name       string // config basename / set id
	WindowName string
	Tabs       []lib.TabSpec
}

func tabSetConfigDir() string {
	if d := strings.TrimSpace(os.Getenv(tabSetDirEnv)); d != "" {
		return d
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".", ".config", "iterm2", "tab-set")
	}
	return filepath.Join(home, ".config", "iterm2", "tab-set")
}

func runTabSet(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stdout, strings.TrimSpace(tabSetHelp)+"\n")
		return nil
	}

	// Global help for tab-set itself.
	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Fprint(stdout, strings.TrimSpace(tabSetHelp)+"\n")
		return nil
	}

	cmd := args[0]
	rest := args[1:]
	switch cmd {
	case "list":
		return tabSetList(rest, stdout, stderr)
	case "show":
		return tabSetShow(rest, stdout, stderr)
	case "run":
		return tabSetRun(rest, stdout, stderr)
	case "status":
		return tabSetStatus(rest, stdout, stderr)
	case "stop":
		return tabSetStop(rest, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "tab-set: unknown subcommand %q\n\n%s\n", cmd, strings.TrimSpace(tabSetHelp))
		return errs.NewSilenceExitCode(1)
	}
}

func tabSetList(args []string, stdout, stderr io.Writer) error {
	if err := rejectExtra(args, "list"); err != nil {
		fmt.Fprint(stderr, err.Error()+"\n")
		return errs.NewSilenceExitCode(1)
	}
	dir := tabSetConfigDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(stdout, "0 sets")
			return nil
		}
		fmt.Fprint(stderr, err.Error()+"\n")
		return errs.NewSilenceExitCode(1)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		names = append(names, strings.TrimSuffix(name, ".json"))
	}
	sort.Strings(names)
	if len(names) == 0 {
		fmt.Fprintln(stdout, "0 sets")
		return nil
	}
	for _, name := range names {
		loaded, lerr := loadTabSet(name)
		if lerr != nil {
			fmt.Fprintf(stdout, "%s  (invalid: %v)\n", name, lerr)
			continue
		}
		fmt.Fprintf(stdout, "%s  (%d tabs)\n", name, len(loaded.Tabs))
	}
	return nil
}

func tabSetShow(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, "tab-set show: missing set name\n")
		return errs.NewSilenceExitCode(1)
	}
	name := args[0]
	if err := rejectExtra(args[1:], "show"); err != nil {
		fmt.Fprint(stderr, err.Error()+"\n")
		return errs.NewSilenceExitCode(1)
	}
	loaded, err := loadTabSet(name)
	if err != nil {
		fmt.Fprint(stderr, err.Error()+"\n")
		return errs.NewSilenceExitCode(1)
	}
	printTabSetDetails(stdout, loaded)
	return nil
}

func printTabSetDetails(w io.Writer, loaded *loadedTabSet) {
	fmt.Fprintf(w, "set: %s\n", loaded.Name)
	if loaded.WindowName != "" {
		fmt.Fprintf(w, "window_name: %s\n", loaded.WindowName)
	}
	fmt.Fprintf(w, "tabs: %d\n", len(loaded.Tabs))
	for _, tab := range loaded.Tabs {
		line := fmt.Sprintf("  - id=%s", tab.ID)
		if tab.Name != "" {
			line += fmt.Sprintf(" name=%s", tab.Name)
		}
		if tab.Command != "" {
			line += fmt.Sprintf(" command=%s", tab.Command)
		}
		if tab.Cwd != "" {
			line += fmt.Sprintf(" cwd=%s", tab.Cwd)
		}
		fmt.Fprintln(w, line)
	}
}

func tabSetRun(args []string, stdout, stderr io.Writer) error {
	var dryRun bool
	var newWindow bool
	var noNewWindow bool
	var save bool
	var force bool
	var windowName string
	var tabs []string
	remain, err := lessflags.Bool("--dry-run", &dryRun).
		Bool("-n,--new-window", &newWindow).
		Bool("--no-new-window", &noNewWindow).
		Bool("--save", &save).
		Bool("--force", &force).
		String("--window-name", &windowName).
		StringSlice("--tab", &tabs).
		HelpFunc("-h,--help", func() {}).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if err == lessflags.ErrHelp {
			fmt.Fprint(stdout, strings.TrimSpace(tabSetHelp)+"\n")
			return nil
		}
		fmt.Fprint(stderr, err.Error()+"\n")
		return errs.NewSilenceExitCode(1)
	}
	if newWindow && noNewWindow {
		fmt.Fprint(stderr, "tab-set run: cannot specify both -n/--new-window and --no-new-window (mutually exclusive)\n")
		return errs.NewSilenceExitCode(1)
	}
	if force && !save {
		fmt.Fprint(stderr, "tab-set run: --force requires --save\n")
		return errs.NewSilenceExitCode(1)
	}
	if save && len(tabs) == 0 {
		fmt.Fprint(stderr, "tab-set run: --save requires at least one --tab\n")
		return errs.NewSilenceExitCode(1)
	}
	if save && (newWindow || noNewWindow) {
		fmt.Fprint(stderr, "tab-set run: --save cannot be used with -n/--new-window or --no-new-window\n")
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) == 0 {
		fmt.Fprint(stderr, "tab-set run: missing set name\n")
		return errs.NewSilenceExitCode(1)
	}
	if len(remain) > 1 {
		fmt.Fprintf(stderr, "tab-set run: unexpected arguments: %s\n", strings.Join(remain[1:], " "))
		return errs.NewSilenceExitCode(1)
	}
	name := remain[0]

	// Ad-hoc mode: ≥1 --tab → do not read config JSON.
	if len(tabs) > 0 {
		parsed, perr := parseAdHocTabs(tabs)
		if perr != nil {
			fmt.Fprint(stderr, perr.Error()+"\n")
			return errs.NewSilenceExitCode(1)
		}
		loaded := &loadedTabSet{
			Name:       name,
			WindowName: windowName,
			Tabs:       parsed,
		}
		if save {
			return tabSetSave(loaded, force, dryRun, stdout, stderr)
		}
		return tabSetRunLoaded(loaded, dryRun, newWindow, noNewWindow, stdout, stderr)
	}

	// Config mode: load <name>.json (--window-name applies only in ad-hoc mode).
	loaded, err := loadTabSet(name)
	if err != nil {
		fmt.Fprint(stderr, err.Error()+"\n")
		return errs.NewSilenceExitCode(1)
	}
	return tabSetRunLoaded(loaded, dryRun, newWindow, noNewWindow, stdout, stderr)
}

func tabSetRunLoaded(loaded *loadedTabSet, dryRun, newWindow, noNewWindow bool, stdout, stderr io.Writer) error {
	mode := lib.TabSetRunSmart
	modeName := "smart"
	if newWindow {
		mode = lib.TabSetRunNewWindow
		modeName = "new-window"
	} else if noNewWindow {
		mode = lib.TabSetRunNoNewWindow
		modeName = "no-new-window"
	}

	if dryRun {
		printDryRunPlan(stdout, loaded, modeName)
		return nil
	}

	spec := toTabSetSpec(loaded)
	cfg := productionTabSetConfig()
	result, err := lib.RunTabSet(spec, lib.TabSetRunOptions{Mode: mode}, cfg)
	if err != nil {
		fmt.Fprint(stderr, err.Error()+"\n")
		return errs.NewSilenceExitCode(1)
	}
	printRunResult(stdout, loaded.Name, result)
	return nil
}

// parseAdHocTabs parses repeatable --tab specs into TabSpecs with defaults.
func parseAdHocTabs(specs []string) ([]lib.TabSpec, error) {
	seen := map[string]bool{}
	tabs := make([]lib.TabSpec, 0, len(specs))
	for i, raw := range specs {
		tab, err := parseTabSpec(raw, i+1)
		if err != nil {
			return nil, err
		}
		if seen[tab.ID] {
			return nil, fmt.Errorf("tab-set run: duplicate tab id %q", tab.ID)
		}
		seen[tab.ID] = true
		tabs = append(tabs, tab)
	}
	return tabs, nil
}

// parseTabSpec parses: spaces [ spaces props spaces ] spaces command
// props are key=value comma-separated (keys: id, name, cwd).
func parseTabSpec(raw string, index int) (lib.TabSpec, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return lib.TabSpec{}, fmt.Errorf("tab-set run: empty --tab command")
	}

	var id, name, cwd, command string
	if strings.HasPrefix(s, "[") {
		end := strings.Index(s, "]")
		if end < 0 {
			return lib.TabSpec{}, fmt.Errorf("tab-set run: invalid tab props: missing closing ]")
		}
		propsBody := s[1:end]
		command = strings.TrimSpace(s[end+1:])
		props, err := parseTabProps(propsBody)
		if err != nil {
			return lib.TabSpec{}, err
		}
		id = props["id"]
		name = props["name"]
		cwd = props["cwd"]
	} else {
		command = s
	}

	if command == "" {
		return lib.TabSpec{}, fmt.Errorf("tab-set run: empty --tab command")
	}
	if id == "" {
		id = fmt.Sprintf("tab-%d", index)
	}
	if name == "" {
		name = id
	}
	return lib.TabSpec{
		ID:      id,
		Name:    name,
		Command: command,
		Cwd:     cwd,
	}, nil
}

func parseTabProps(body string) (map[string]string, error) {
	body = strings.TrimSpace(body)
	result := map[string]string{}
	if body == "" {
		return result, nil
	}
	parts := strings.Split(body, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		eq := strings.Index(part, "=")
		if eq < 0 {
			return nil, fmt.Errorf("tab-set run: invalid tab props: expected key=value, got %q", part)
		}
		key := strings.TrimSpace(part[:eq])
		val := strings.TrimSpace(part[eq+1:])
		if key == "" {
			return nil, fmt.Errorf("tab-set run: invalid tab props: empty key")
		}
		switch key {
		case "id", "name", "cwd":
			result[key] = val
		default:
			return nil, fmt.Errorf("tab-set run: invalid tab props: unknown key %q", key)
		}
	}
	return result, nil
}

func tabSetSave(loaded *loadedTabSet, force, dryRun bool, stdout, stderr io.Writer) error {
	dir := tabSetConfigDir()
	path := filepath.Join(dir, loaded.Name+".json")

	newFile := loadedToTabSetFile(loaded)

	var existing *tabSetFile
	data, err := os.ReadFile(path)
	exists := err == nil
	if exists {
		var file tabSetFile
		if jerr := json.Unmarshal(data, &file); jerr != nil {
			// Treat unreadable existing as overwrite target without structured diff.
			existing = nil
		} else {
			existing = &file
		}
	} else if !os.IsNotExist(err) {
		fmt.Fprint(stderr, err.Error()+"\n")
		return errs.NewSilenceExitCode(1)
	}

	if !exists {
		if dryRun {
			fmt.Fprintf(stdout, "dry-run save plan for tab-set %q\n", loaded.Name)
			fmt.Fprintln(stdout, "would create new config file")
			printSavePlanTabs(stdout, newFile)
			return nil
		}
		if err := writeTabSetFile(path, newFile); err != nil {
			fmt.Fprint(stderr, err.Error()+"\n")
			return errs.NewSilenceExitCode(1)
		}
		fmt.Fprintf(stdout, "saved tab-set %q (%d tabs)\n", loaded.Name, len(newFile.Tabs))
		return nil
	}

	// Existing file: print nice diff always, then confirm / force / dry-run.
	printTabSetDiff(stdout, loaded.Name, existing, newFile)

	if dryRun {
		fmt.Fprintln(stdout, "dry-run: would overwrite (no write)")
		return nil
	}

	if !force {
		if !terminal.IsStdinTerminal() {
			fmt.Fprint(stderr, "tab-set run: config exists; refuse to overwrite without --force (non-interactive / non-TTY)\n")
			return errs.NewSilenceExitCode(1)
		}
		fmt.Fprint(stdout, "Overwrite? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		ans := strings.TrimSpace(strings.ToLower(line))
		if ans != "y" && ans != "yes" {
			fmt.Fprint(stderr, "tab-set run: overwrite declined\n")
			return errs.NewSilenceExitCode(1)
		}
	}

	if err := writeTabSetFile(path, newFile); err != nil {
		fmt.Fprint(stderr, err.Error()+"\n")
		return errs.NewSilenceExitCode(1)
	}
	fmt.Fprintf(stdout, "saved tab-set %q (%d tabs)\n", loaded.Name, len(newFile.Tabs))
	return nil
}

func loadedToTabSetFile(loaded *loadedTabSet) *tabSetFile {
	tabs := make([]tabSetTab, 0, len(loaded.Tabs))
	for _, t := range loaded.Tabs {
		tabs = append(tabs, tabSetTab{
			ID:      t.ID,
			Name:    t.Name,
			Command: t.Command,
			Cwd:     t.Cwd,
		})
	}
	return &tabSetFile{
		Version:    1,
		WindowName: loaded.WindowName,
		Tabs:       tabs,
	}
}

func writeTabSetFile(path string, file *tabSetFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

func printSavePlanTabs(w io.Writer, file *tabSetFile) {
	if file.WindowName != "" {
		fmt.Fprintf(w, "window_name: %s\n", file.WindowName)
	}
	fmt.Fprintf(w, "tabs: %d\n", len(file.Tabs))
	for _, tab := range file.Tabs {
		fmt.Fprintf(w, "  - id=%s command=%s\n", tab.ID, tab.Command)
	}
}

// printTabSetDiff compares by id into unchanged/modified/added/deleted buckets.
func printTabSetDiff(w io.Writer, setName string, oldFile, newFile *tabSetFile) {
	fmt.Fprintf(w, "diff for tab-set %q\n", setName)

	oldWin := ""
	if oldFile != nil {
		oldWin = oldFile.WindowName
	}
	newWin := ""
	if newFile != nil {
		newWin = newFile.WindowName
	}
	if oldWin != newWin {
		fmt.Fprintln(w, "window_name: modified")
		fmt.Fprintf(w, "  - %s\n", oldWin)
		fmt.Fprintf(w, "  + %s\n", newWin)
	} else {
		fmt.Fprintln(w, "window_name: unchanged")
	}

	oldByID := map[string]tabSetTab{}
	if oldFile != nil {
		for _, t := range oldFile.Tabs {
			id := t.ID
			if id == "" {
				id = t.Name
			}
			oldByID[id] = t
		}
	}
	newByID := map[string]tabSetTab{}
	newOrder := []string{}
	if newFile != nil {
		for _, t := range newFile.Tabs {
			id := t.ID
			if id == "" {
				id = t.Name
			}
			newByID[id] = t
			newOrder = append(newOrder, id)
		}
	}

	var unchanged, modified, added, deleted []string
	for _, id := range newOrder {
		nt := newByID[id]
		ot, ok := oldByID[id]
		if !ok {
			added = append(added, id)
			continue
		}
		if tabContentEqual(ot, nt) {
			unchanged = append(unchanged, id)
		} else {
			modified = append(modified, id)
		}
	}
	// deleted: preserve old file order
	if oldFile != nil {
		for _, t := range oldFile.Tabs {
			id := t.ID
			if id == "" {
				id = t.Name
			}
			if _, ok := newByID[id]; !ok {
				deleted = append(deleted, id)
			}
		}
	}

	fmt.Fprintln(w, "unchanged:")
	if len(unchanged) == 0 {
		fmt.Fprintln(w, "  (none)")
	} else {
		for _, id := range unchanged {
			fmt.Fprintf(w, "  - %s\n", id)
		}
	}
	fmt.Fprintln(w, "modified:")
	if len(modified) == 0 {
		fmt.Fprintln(w, "  (none)")
	} else {
		for _, id := range modified {
			ot := oldByID[id]
			nt := newByID[id]
			fmt.Fprintf(w, "  - %s\n", id)
			if ot.Command != nt.Command {
				fmt.Fprintln(w, "    command:")
				fmt.Fprintf(w, "      - %s\n", ot.Command)
				fmt.Fprintf(w, "      + %s\n", nt.Command)
			}
			if ot.Name != nt.Name {
				fmt.Fprintln(w, "    name:")
				fmt.Fprintf(w, "      - %s\n", ot.Name)
				fmt.Fprintf(w, "      + %s\n", nt.Name)
			}
			if ot.Cwd != nt.Cwd {
				fmt.Fprintln(w, "    cwd:")
				fmt.Fprintf(w, "      - %s\n", ot.Cwd)
				fmt.Fprintf(w, "      + %s\n", nt.Cwd)
			}
		}
	}
	fmt.Fprintln(w, "added:")
	if len(added) == 0 {
		fmt.Fprintln(w, "  (none)")
	} else {
		for _, id := range added {
			fmt.Fprintf(w, "  - %s\n", id)
		}
	}
	fmt.Fprintln(w, "deleted:")
	if len(deleted) == 0 {
		fmt.Fprintln(w, "  (none)")
	} else {
		for _, id := range deleted {
			fmt.Fprintf(w, "  - %s\n", id)
		}
	}
}

func tabContentEqual(a, b tabSetTab) bool {
	// Compare semantic fields used for identity of tab content.
	aName, bName := a.Name, b.Name
	if aName == "" {
		aName = a.ID
	}
	if bName == "" {
		bName = b.ID
	}
	return a.Command == b.Command && aName == bName && a.Cwd == b.Cwd
}

func printDryRunPlan(w io.Writer, loaded *loadedTabSet, modeName string) {
	fmt.Fprintf(w, "dry-run plan for tab-set %q\n", loaded.Name)
	fmt.Fprintf(w, "mode: %s\n", modeName)
	if loaded.WindowName != "" {
		fmt.Fprintf(w, "window_name: %s\n", loaded.WindowName)
	}
	fmt.Fprintf(w, "would run %d tabs:\n", len(loaded.Tabs))
	for _, tab := range loaded.Tabs {
		fmt.Fprintf(w, "  - %s: %s\n", tab.ID, tab.Command)
	}
}

func printRunResult(w io.Writer, setName string, result *lib.TabSetRunResult) {
	if result == nil {
		return
	}
	fmt.Fprintf(w, "tab-set %q\n", setName)
	if result.CreatedWindow {
		fmt.Fprintln(w, "created new window")
	}
	if result.FocusedWindow != "" {
		fmt.Fprintf(w, "focused window: %s\n", result.FocusedWindow)
	}
	if result.Warning != "" {
		fmt.Fprintf(w, "warning: %s\n", result.Warning)
	}
	for _, tr := range result.Tabs {
		fmt.Fprintf(w, "  tab %s: %s\n", tr.TabID, tr.Action)
	}
}

func tabSetStatus(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, "tab-set status: missing set name\n")
		return errs.NewSilenceExitCode(1)
	}
	name := args[0]
	if err := rejectExtra(args[1:], "status"); err != nil {
		fmt.Fprint(stderr, err.Error()+"\n")
		return errs.NewSilenceExitCode(1)
	}
	loaded, err := loadTabSet(name)
	if err != nil {
		fmt.Fprint(stderr, err.Error()+"\n")
		return errs.NewSilenceExitCode(1)
	}
	st, err := lib.StatusTabSet(toTabSetSpec(loaded), productionTabSetConfig())
	if err != nil {
		fmt.Fprint(stderr, err.Error()+"\n")
		return errs.NewSilenceExitCode(1)
	}
	fmt.Fprintf(stdout, "set: %s\n", st.SetName)
	fmt.Fprintf(stdout, "instances: %d\n", st.Instances)
	if st.WindowID != "" {
		fmt.Fprintf(stdout, "window: %s\n", st.WindowID)
	}
	if st.Warning != "" {
		fmt.Fprintf(stdout, "warning: %s\n", st.Warning)
	}
	for _, e := range st.Tabs {
		fmt.Fprintf(stdout, "  tab %s: %s\n", e.TabID, e.State)
	}
	return nil
}

func tabSetStop(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, "tab-set stop: missing set name\n")
		return errs.NewSilenceExitCode(1)
	}
	name := args[0]
	if err := rejectExtra(args[1:], "stop"); err != nil {
		fmt.Fprint(stderr, err.Error()+"\n")
		return errs.NewSilenceExitCode(1)
	}
	// Ensure config exists (consistent UX with show/run).
	if _, err := loadTabSet(name); err != nil {
		fmt.Fprint(stderr, err.Error()+"\n")
		return errs.NewSilenceExitCode(1)
	}
	res, err := lib.StopTabSet(name, productionTabSetConfig())
	if err != nil {
		fmt.Fprint(stderr, err.Error()+"\n")
		return errs.NewSilenceExitCode(1)
	}
	if res.Warning != "" {
		fmt.Fprintf(stdout, "warning: %s\n", res.Warning)
	}
	fmt.Fprintf(stdout, "closed windows: %d\n", res.ClosedWindows)
	fmt.Fprintf(stdout, "closed tabs: %d\n", res.ClosedTabs)
	return nil
}

func loadTabSet(name string) (*loadedTabSet, error) {
	if name == "" {
		return nil, fmt.Errorf("tab-set: missing set name")
	}
	path := filepath.Join(tabSetConfigDir(), name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("tab-set %q not found (looked in %s)", name, tabSetConfigDir())
		}
		return nil, fmt.Errorf("tab-set %q: %w", name, err)
	}
	var file tabSetFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("tab-set %q: invalid JSON: %w", name, err)
	}
	if file.Version != 1 {
		return nil, fmt.Errorf("tab-set %q: unsupported version %d (want 1)", name, file.Version)
	}
	if len(file.Tabs) == 0 {
		return nil, fmt.Errorf("tab-set %q: tabs must not be empty", name)
	}
	seen := map[string]bool{}
	tabs := make([]lib.TabSpec, 0, len(file.Tabs))
	for i, t := range file.Tabs {
		id := strings.TrimSpace(t.ID)
		if id == "" {
			id = strings.TrimSpace(t.Name)
		}
		if id == "" {
			return nil, fmt.Errorf("tab-set %q: tab %d: id or name is required", name, i)
		}
		if seen[id] {
			return nil, fmt.Errorf("tab-set %q: duplicate tab id %q", name, id)
		}
		seen[id] = true
		cmd := strings.TrimSpace(t.Command)
		if cmd == "" {
			return nil, fmt.Errorf("tab-set %q: tab %q: command is required", name, id)
		}
		tabs = append(tabs, lib.TabSpec{
			ID:      id,
			Name:    t.Name,
			Command: cmd,
			Cwd:     t.Cwd,
		})
	}
	return &loadedTabSet{
		Name:       name,
		WindowName: file.WindowName,
		Tabs:       tabs,
	}, nil
}

func toTabSetSpec(loaded *loadedTabSet) lib.TabSetSpec {
	return lib.TabSetSpec{
		Name:       loaded.Name,
		WindowName: loaded.WindowName,
		Tabs:       loaded.Tabs,
	}
}

// productionTabSetConfig returns a TabSetConfig that uses library defaults
// (Find via BuildTabSetFindScript + osascript, Exec via osascript).
func productionTabSetConfig() *lib.TabSetConfig {
	// nil fields → lib.normalizeTabSetConfig fills Find/Busy/Exec defaults.
	return &lib.TabSetConfig{}
}

func rejectExtra(args []string, cmd string) error {
	if len(args) == 0 {
		return nil
	}
	return fmt.Errorf("tab-set %s: unexpected arguments: %s", cmd, strings.Join(args, " "))
}
