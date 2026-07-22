package tasks

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/xhd2015/kool/pkgs/errs"
)

// resolveBackend picks local | iterm2 from explicit backend or auto heuristic.
func resolveBackend(requested string, steps []PlanStep) (string, error) {
	b := strings.ToLower(strings.TrimSpace(requested))
	if b == "" {
		b = "auto"
	}
	switch b {
	case "local", "iterm2":
		return b, nil
	case "auto":
		// 1 non-background leaf with a command → local; multi or any BG → iterm2 on darwin else local
		var leaves []PlanStep
		anyBG := false
		for _, s := range steps {
			if s.Command == "" {
				continue // composite bookkeeping
			}
			leaves = append(leaves, s)
			if s.BG {
				anyBG = true
			}
		}
		if len(leaves) == 1 && !anyBG {
			return "local", nil
		}
		if runtime.GOOS == "darwin" {
			return "iterm2", nil
		}
		return "local", nil
	default:
		return "", fmt.Errorf("invalid --backend %q (want auto|local|iterm2)", requested)
	}
}

func printDryRunPlan(ws *Workspace, root *Task, steps []PlanStep, backend string) {
	if backend == "iterm2" {
		fmt.Printf("dry-run iterm2 tab plan\n")
		fmt.Printf("workspace: %s\n", ws.Root)
		fmt.Printf("root task: %s\n", root.Label)
		// tab-like mapping for leaf steps (and composites listed without tabs)
		tabN := 0
		for _, s := range steps {
			if s.Command == "" {
				fmt.Printf("  [composite] %s\n", s.Label)
				continue
			}
			tabN++
			bg := ""
			if s.BG {
				bg = " [background]"
			}
			fmt.Printf("  tab %d: [%s] %s%s\n", tabN, s.Kind, s.Label, bg)
			fmt.Printf("     command: %s\n", s.Command)
			if s.Cwd != "" {
				fmt.Printf("     cwd: %s\n", s.Cwd)
			}
		}
		fmt.Printf("tabs: %d\n", tabN)
		return
	}

	fmt.Printf("dry-run plan\n")
	fmt.Printf("workspace: %s\n", ws.Root)
	fmt.Printf("root task: %s\n", root.Label)
	fmt.Printf("steps: %d\n", len(steps))
	for i, s := range steps {
		bg := ""
		if s.BG {
			bg = " [background]"
		}
		fmt.Printf("  %d. [%s] %s%s\n", i+1, s.Kind, s.Label, bg)
		if s.Command != "" {
			fmt.Printf("     command: %s\n", s.Command)
		}
		if s.Cwd != "" {
			fmt.Printf("     cwd: %s\n", s.Cwd)
		}
	}
}

// runLocal executes leaf plan steps sequentially via shell or process.
func runLocal(steps []PlanStep, byLabel map[string]*Task, workspaceRoot string) error {
	var lastCode int
	ran := 0
	for _, s := range steps {
		if s.Command == "" {
			continue
		}
		t := byLabel[s.Label]
		if t == nil {
			return fmt.Errorf("internal: missing task %q for plan step", s.Label)
		}
		cmd, err := buildExecCmd(t, workspaceRoot)
		if err != nil {
			return err
		}
		if s.Cwd != "" {
			cmd.Dir = s.Cwd
		} else {
			cmd.Dir = workspaceRoot
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		runErr := cmd.Run()
		ran++
		if runErr != nil {
			if ee, ok := runErr.(*exec.ExitError); ok {
				code := ee.ExitCode()
				if code == 0 {
					code = 1
				}
				return errs.NewSilenceExitCode(code)
			}
			return fmt.Errorf("run task %q: %w", s.Label, runErr)
		}
		lastCode = 0
	}
	if ran == 0 {
		// composite with only deps that had no commands?
		_ = lastCode
	}
	return nil
}

func buildExecCmd(t *Task, workspaceRoot string) (*exec.Cmd, error) {
	kind := t.Kind()
	cwdHint := workspaceRoot

	// Expand command and args individually for process argv.
	cmdStr, err := ExpandVars(t.Command, workspaceRoot)
	if err != nil {
		return nil, err
	}
	args := make([]string, len(t.Args))
	for i, a := range t.Args {
		args[i], err = ExpandVars(a, workspaceRoot)
		if err != nil {
			return nil, err
		}
	}

	if kind == "process" {
		if cmdStr == "" {
			return nil, fmt.Errorf("process task %q has empty command", t.Label)
		}
		return exec.Command(cmdStr, args...), nil
	}

	// shell (default): sh -c with full command line
	line := cmdStr
	if len(args) > 0 {
		parts := make([]string, 0, 1+len(args))
		parts = append(parts, cmdStr)
		parts = append(parts, args...)
		line = strings.Join(parts, " ")
	}
	if line == "" {
		return nil, fmt.Errorf("shell task %q has empty command", t.Label)
	}
	_ = cwdHint
	return exec.Command("sh", "-c", line), nil
}
