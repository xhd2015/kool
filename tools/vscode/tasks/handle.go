package tasks

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

const usage = `kool vscode tasks — discover and plan VS Code workspace tasks

Usage:
  kool vscode tasks list [--dir <path>] [--json]
  kool vscode tasks find <query> [--dir <path>]
  kool vscode tasks show <label> [--dir <path>]
  kool vscode tasks run  <label> [--dir <path>] [--dry-run]
      [--backend=auto|local|iterm2]
      [-n|--new-window] [--no-new-window]
  kool vscode tasks -h|--help

Discovers <workspace>/.vscode/tasks.json (JSONC) by walking up from --dir or cwd.

Commands:
  list   list all tasks (LABEL, TYPE, BG, DEPS, COMMAND)
  find   case-insensitive substring search on labels
  show   show one task (exact label, else unique CI substring)
  run    execute a task (local/iterm2/auto) or print plan with --dry-run

Flags:
  --dir <path>              start directory for workspace discovery
  --json                    machine-readable list output
  --dry-run                 expand dependsOn and variables; do not execute
  --backend auto|local|iterm2   run backend (default: auto)
  -n, --new-window          open in a new iTerm2 window (iterm2 only)
  --no-new-window           reuse existing iTerm2 window (iterm2 only)
  -h, --help                show this help
`

// Handle is the entry for `kool vscode tasks …`.
func Handle(args []string) error {
	if len(args) == 0 {
		fmt.Print(usage)
		return nil
	}

	// global help
	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		fmt.Print(usage)
		return nil
	}

	sub := args[0]
	rest := args[1:]

	switch sub {
	case "list":
		return cmdList(rest)
	case "find":
		return cmdFind(rest)
	case "show":
		return cmdShow(rest)
	case "run":
		return cmdRun(rest)
	default:
		return fmt.Errorf("unrecognized tasks subcommand: %s\n\n%s", sub, usage)
	}
}

type commonFlags struct {
	Dir         string
	JSON        bool
	DryRun      bool
	Help        bool
	Backend     string
	NewWindow   bool
	NoNewWindow bool
	// remaining positional
	Pos []string
}

func parseFlags(args []string) (*commonFlags, error) {
	f := &commonFlags{}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help":
			f.Help = true
		case a == "--json":
			f.JSON = true
		case a == "--dry-run":
			f.DryRun = true
		case a == "-n" || a == "--new-window":
			f.NewWindow = true
		case a == "--no-new-window":
			f.NoNewWindow = true
		case a == "--backend":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--backend requires a value (auto|local|iterm2)")
			}
			i++
			f.Backend = args[i]
		case strings.HasPrefix(a, "--backend="):
			f.Backend = strings.TrimPrefix(a, "--backend=")
		case a == "--dir":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--dir requires a path")
			}
			i++
			f.Dir = args[i]
		case strings.HasPrefix(a, "--dir="):
			f.Dir = strings.TrimPrefix(a, "--dir=")
		case strings.HasPrefix(a, "-"):
			return nil, fmt.Errorf("unrecognized flag: %s", a)
		default:
			f.Pos = append(f.Pos, a)
		}
	}
	return f, nil
}

func loadFromFlags(f *commonFlags) (*Workspace, *File, error) {
	ws, err := FindWorkspace(f.Dir)
	if err != nil {
		return nil, nil, err
	}
	file, err := LoadTasks(ws)
	if err != nil {
		return nil, nil, err
	}
	return ws, file, nil
}

func cmdList(args []string) error {
	f, err := parseFlags(args)
	if err != nil {
		return err
	}
	if f.Help {
		fmt.Print(usage)
		return nil
	}
	ws, file, err := loadFromFlags(f)
	if err != nil {
		return err
	}

	// sort by label
	tasks := make([]*Task, 0, len(file.Tasks))
	for i := range file.Tasks {
		tasks = append(tasks, &file.Tasks[i])
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Label < tasks[j].Label
	})

	if f.JSON {
		type row struct {
			Label        string   `json:"label"`
			Type         string   `json:"type"`
			IsBackground bool     `json:"isBackground"`
			DependsOn    []string `json:"dependsOn,omitempty"`
			Command      string   `json:"command,omitempty"`
			Workspace    string   `json:"workspace,omitempty"`
		}
		rows := make([]row, 0, len(tasks))
		for _, t := range tasks {
			rows = append(rows, row{
				Label:        t.Label,
				Type:         t.Kind(),
				IsBackground: t.IsBackground,
				DependsOn:    t.DependsOn,
				Command:      t.CommandPreview(),
				Workspace:    ws.Root,
			})
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(rows); err != nil {
			return err
		}
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "LABEL\tTYPE\tBG\tDEPS\tCOMMAND")
	for _, t := range tasks {
		bg := "no"
		if t.IsBackground {
			bg = "yes"
		}
		deps := ""
		if n := len(t.DependsOn); n > 0 {
			deps = fmt.Sprintf("%d", n)
			if n <= 3 {
				deps = strings.Join(t.DependsOn, ",")
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", t.Label, t.Kind(), bg, deps, t.CommandPreview())
	}
	_ = w.Flush()
	fmt.Printf("\n%d task(s)  workspace: %s\n", len(tasks), ws.Root)
	return nil
}

func cmdFind(args []string) error {
	f, err := parseFlags(args)
	if err != nil {
		return err
	}
	if f.Help {
		fmt.Print(usage)
		return nil
	}
	if len(f.Pos) == 0 {
		return fmt.Errorf("usage: kool vscode tasks find <query> [--dir <path>]")
	}
	query := f.Pos[0]
	ws, file, err := loadFromFlags(f)
	if err != nil {
		return err
	}
	matches := file.FindSubstring(query)
	if len(matches) == 0 {
		return fmt.Errorf("no task matches %q (not found)", query)
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Label < matches[j].Label
	})
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "LABEL\tTYPE\tBG\tCOMMAND")
	for _, t := range matches {
		bg := "no"
		if t.IsBackground {
			bg = "yes"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.Label, t.Kind(), bg, t.CommandPreview())
	}
	_ = w.Flush()
	fmt.Printf("\n%d match(es) in workspace %s\n", len(matches), ws.Root)
	return nil
}

func cmdShow(args []string) error {
	f, err := parseFlags(args)
	if err != nil {
		return err
	}
	if f.Help {
		fmt.Print(usage)
		return nil
	}
	if len(f.Pos) == 0 {
		return fmt.Errorf("usage: kool vscode tasks show <label> [--dir <path>]")
	}
	query := strings.Join(f.Pos, " ")
	// if multiple pos args, join for labels with spaces — but flags already separated.
	// Prefer first positional only when only one; for "Build All" test harness
	// passes as single argv "Build All".
	if len(f.Pos) == 1 {
		query = f.Pos[0]
	}
	ws, file, err := loadFromFlags(f)
	if err != nil {
		return err
	}
	t, err := file.MatchOne(query)
	if err != nil {
		return err
	}
	fmt.Printf("label: %s\n", t.Label)
	fmt.Printf("type: %s\n", t.Kind())
	if t.Command != "" {
		fmt.Printf("command: %s\n", t.Command)
	}
	if len(t.Args) > 0 {
		fmt.Printf("args: %s\n", strings.Join(t.Args, " "))
	}
	if t.Options != nil && t.Options.Cwd != "" {
		fmt.Printf("cwd: %s\n", t.Options.Cwd)
	}
	fmt.Printf("isBackground: %v\n", t.IsBackground)
	if len(t.DependsOn) > 0 {
		fmt.Printf("dependsOn: %s\n", strings.Join(t.DependsOn, ", "))
	}
	fmt.Printf("workspace: %s\n", ws.Root)
	fmt.Printf("tasks.json: %s\n", ws.TasksPath)
	return nil
}

func cmdRun(args []string) error {
	f, err := parseFlags(args)
	if err != nil {
		return err
	}
	if f.Help {
		fmt.Print(usage)
		return nil
	}
	if len(f.Pos) == 0 {
		return fmt.Errorf("usage: kool vscode tasks run <label> [--dir <path>] [--dry-run] [--backend=auto|local|iterm2]")
	}
	query := f.Pos[0]
	if len(f.Pos) > 1 {
		query = strings.Join(f.Pos, " ")
	}

	// Window flags: always mutually exclusive.
	if f.NewWindow && f.NoNewWindow {
		return fmt.Errorf("-n/--new-window and --no-new-window are mutually exclusive")
	}

	ws, file, err := loadFromFlags(f)
	if err != nil {
		return err
	}
	t, err := file.MatchOne(query)
	if err != nil {
		return err
	}
	byLabel := file.ByLabel()
	steps, err := BuildPlan(t, byLabel, ws.Root)
	if err != nil {
		return err
	}

	backend, err := resolveBackend(f.Backend, steps)
	if err != nil {
		return err
	}

	// Window flags only valid for iterm2 (not local).
	if f.NewWindow || f.NoNewWindow {
		if backend == "local" {
			return fmt.Errorf("-n/--new-window and --no-new-window are not supported with --backend=local (iterm2 only)")
		}
	}

	if f.DryRun {
		printDryRunPlan(ws, t, steps, backend)
		return nil
	}

	switch backend {
	case "local":
		return runLocal(steps, byLabel, ws.Root)
	case "iterm2":
		// Live iterm2: map leaves → TabSetSpec and RunTabSet (or CI mock seam).
		// Fail closed on error — never silent multi-leaf local fallback.
		return runITerm2(t, steps, ws.Root, f.NewWindow, f.NoNewWindow)
	default:
		return fmt.Errorf("internal: unknown backend %q", backend)
	}
}
