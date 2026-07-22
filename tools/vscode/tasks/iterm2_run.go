package tasks

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	lib "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// runITerm2 maps plan leaves to an ephemeral TabSetSpec and invokes RunTabSet
// (or the KOOL_VSCODE_TASKS_ITERM2_MOCK test seam). Fail closed: never fall back
// to sequential local multi-leaf execution.
func runITerm2(root *Task, steps []PlanStep, workspaceRoot string, newWindow, noNewWindow bool) error {
	spec := buildTabSetSpec(root, steps, workspaceRoot)
	modeName, mode := tabSetMode(newWindow, noNewWindow)

	if os.Getenv("KOOL_VSCODE_TASKS_ITERM2_MOCK") == "1" {
		if err := writeITerm2MockCall(modeName, spec); err != nil {
			return err
		}
		if os.Getenv("KOOL_VSCODE_TASKS_ITERM2_MOCK_ERR") == "1" {
			return fmt.Errorf("iterm2 mock error (KOOL_VSCODE_TASKS_ITERM2_MOCK_ERR=1)")
		}
		fmt.Printf("iterm2: launched tab set %q (%d tabs, mode %s)\n", spec.Name, len(spec.Tabs), modeName)
		return nil
	}

	cfg := &lib.TabSetConfig{} // production hooks (Find/Busy/Exec) via library defaults
	result, err := lib.RunTabSet(spec, lib.TabSetRunOptions{Mode: mode}, cfg)
	if err != nil {
		return fmt.Errorf("iterm2 RunTabSet: %w", err)
	}
	if result != nil && result.Warning != "" {
		fmt.Fprintf(os.Stderr, "warning: %s\n", result.Warning)
	}
	fmt.Printf("iterm2: launched tab set %q (%d tabs, mode %s)\n", spec.Name, len(spec.Tabs), modeName)
	return nil
}

func tabSetMode(newWindow, noNewWindow bool) (string, lib.TabSetRunMode) {
	if newWindow {
		return "new-window", lib.TabSetRunNewWindow
	}
	if noNewWindow {
		return "no-new-window", lib.TabSetRunNoNewWindow
	}
	return "smart", lib.TabSetRunSmart
}

func buildTabSetSpec(root *Task, steps []PlanStep, workspaceRoot string) lib.TabSetSpec {
	rootLabel := ""
	if root != nil {
		rootLabel = root.Label
	}
	name := "vscode-tasks-" + slugify(rootLabel)
	if rootLabel == "" {
		name = "vscode-tasks"
	}

	tabs := make([]lib.TabSpec, 0, len(steps))
	for _, s := range steps {
		if s.Command == "" {
			continue // composite bookkeeping only
		}
		cwd := s.Cwd
		if cwd == "" {
			cwd = workspaceRoot
		}
		tabs = append(tabs, lib.TabSpec{
			ID:      slugify(s.Label),
			Name:    s.Label,
			Command: s.Command,
			Cwd:     cwd,
		})
	}

	return lib.TabSetSpec{
		Name:       name,
		WindowName: rootLabel,
		Tabs:       tabs,
	}
}

// slugify produces a stable lowercase [a-z0-9-] id from a task label.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "tab"
	}
	s = nonSlugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	if s == "" {
		return "tab"
	}
	return s
}

var nonSlugRe = regexp.MustCompile(`[^a-z0-9]+`)

// mockCallJSON is the CI-visible record of an intended RunTabSet call.
type mockCallJSON struct {
	Mode string `json:"mode"`
	Spec struct {
		Name       string `json:"name"`
		WindowName string `json:"windowName"`
		Tabs       []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Command string `json:"command"`
			Cwd     string `json:"cwd"`
		} `json:"tabs"`
	} `json:"spec"`
}

func writeITerm2MockCall(mode string, spec lib.TabSetSpec) error {
	out := os.Getenv("KOOL_VSCODE_TASKS_ITERM2_MOCK_OUT")
	if out == "" {
		// still succeed without a path (nothing to assert)
		return nil
	}
	call := mockCallJSON{Mode: mode}
	call.Spec.Name = spec.Name
	call.Spec.WindowName = spec.WindowName
	for _, t := range spec.Tabs {
		call.Spec.Tabs = append(call.Spec.Tabs, struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Command string `json:"command"`
			Cwd     string `json:"cwd"`
		}{
			ID:      t.ID,
			Name:    t.Name,
			Command: t.Command,
			Cwd:     t.Cwd,
		})
	}
	data, err := json.MarshalIndent(call, "", "  ")
	if err != nil {
		return fmt.Errorf("iterm2 mock encode: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(out, data, 0o644); err != nil {
		return fmt.Errorf("iterm2 mock write %s: %w", out, err)
	}
	return nil
}
