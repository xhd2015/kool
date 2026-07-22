package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var varPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// ExpandVars expands ${workspaceFolder}, ${workspaceFolderBasename}, ${env:NAME}.
// Unknown ${…} tokens return an error.
func ExpandVars(s string, workspaceRoot string) (string, error) {
	var firstErr error
	out := varPattern.ReplaceAllStringFunc(s, func(match string) string {
		if firstErr != nil {
			return match
		}
		inner := match[2 : len(match)-1] // strip ${ }
		switch {
		case inner == "workspaceFolder":
			return workspaceRoot
		case inner == "workspaceFolderBasename":
			return filepath.Base(workspaceRoot)
		case strings.HasPrefix(inner, "env:"):
			name := strings.TrimPrefix(inner, "env:")
			return os.Getenv(name)
		default:
			firstErr = fmt.Errorf("unresolved variable ${%s}", inner)
			return match
		}
	})
	if firstErr != nil {
		return "", firstErr
	}
	return out, nil
}

// PlanStep is one step in a dry-run plan.
type PlanStep struct {
	Label   string
	Kind    string
	Command string
	Cwd     string
	BG      bool
}

// BuildPlan expands dependsOn DAG (deps first, then root) and expands vars.
// Detects cycles and missing dependencies.
func BuildPlan(root *Task, byLabel map[string]*Task, workspaceRoot string) ([]PlanStep, error) {
	var steps []PlanStep
	visiting := map[string]bool{}
	done := map[string]bool{}

	var visit func(t *Task) error
	visit = func(t *Task) error {
		if t == nil {
			return fmt.Errorf("nil task")
		}
		if done[t.Label] {
			return nil
		}
		if visiting[t.Label] {
			return fmt.Errorf("dependsOn cycle detected involving task %q", t.Label)
		}
		visiting[t.Label] = true
		for _, dep := range t.DependsOn {
			d, ok := byLabel[dep]
			if !ok {
				return fmt.Errorf("missing dependency %q required by task %q", dep, t.Label)
			}
			if err := visit(d); err != nil {
				return err
			}
		}
		// expand command/args/cwd
		cmd := t.CommandPreview()
		var err error
		if cmd != "" {
			cmd, err = ExpandVars(cmd, workspaceRoot)
			if err != nil {
				return err
			}
		}
		cwd := ""
		if t.Options != nil {
			cwd = t.Options.Cwd
		}
		if cwd != "" {
			cwd, err = ExpandVars(cwd, workspaceRoot)
			if err != nil {
				return err
			}
		}
		steps = append(steps, PlanStep{
			Label:   t.Label,
			Kind:    t.Kind(),
			Command: cmd,
			Cwd:     cwd,
			BG:      t.IsBackground,
		})
		delete(visiting, t.Label)
		done[t.Label] = true
		return nil
	}

	if err := visit(root); err != nil {
		return nil, err
	}
	return steps, nil
}
