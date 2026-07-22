package tasks

import (
	"encoding/json"
	"fmt"
	"strings"
)

// File is the top-level tasks.json structure.
type File struct {
	Version string `json:"version"`
	Tasks   []Task `json:"tasks"`
}

// Task is a VS Code task entry.
type Task struct {
	Label        string          `json:"label"`
	Type         string          `json:"type"`
	Command      string          `json:"command"`
	Args         []string        `json:"args"`
	IsBackground bool            `json:"isBackground"`
	Options      *TaskOptions    `json:"options"`
	DependsOnRaw json.RawMessage `json:"dependsOn"`
	DependsOn    []string        `json:"-"`
	Group        json.RawMessage `json:"group"`
}

// TaskOptions holds options.cwd etc.
type TaskOptions struct {
	Cwd string `json:"cwd"`
}

// Kind returns shell | process | composite for display.
func (t *Task) Kind() string {
	if t.Type != "" {
		return t.Type
	}
	if t.Command == "" && len(t.DependsOn) > 0 {
		return "composite"
	}
	if t.Command != "" {
		return "shell"
	}
	if len(t.DependsOn) > 0 {
		return "composite"
	}
	return ""
}

// CommandPreview builds a short command line for list/show.
func (t *Task) CommandPreview() string {
	if t.Command == "" {
		return ""
	}
	if len(t.Args) == 0 {
		return t.Command
	}
	parts := make([]string, 0, 1+len(t.Args))
	parts = append(parts, t.Command)
	parts = append(parts, t.Args...)
	return strings.Join(parts, " ")
}

// ParseTasksJSON parses JSONC tasks content.
func ParseTasksJSON(data []byte) (*File, error) {
	var f File
	if err := UnmarshalJSONC(data, &f); err != nil {
		return nil, err
	}
	for i := range f.Tasks {
		f.Tasks[i].DependsOn = parseDependsOn(f.Tasks[i].DependsOnRaw)
	}
	return &f, nil
}

// ByLabel returns a map label -> *Task (first wins).
func (f *File) ByLabel() map[string]*Task {
	m := make(map[string]*Task, len(f.Tasks))
	for i := range f.Tasks {
		t := &f.Tasks[i]
		if t.Label == "" {
			continue
		}
		if _, ok := m[t.Label]; !ok {
			m[t.Label] = t
		}
	}
	return m
}

// FindSubstring returns tasks whose labels contain query (case-insensitive).
func (f *File) FindSubstring(query string) []*Task {
	q := strings.ToLower(query)
	var out []*Task
	for i := range f.Tasks {
		t := &f.Tasks[i]
		if strings.Contains(strings.ToLower(t.Label), q) {
			out = append(out, t)
		}
	}
	return out
}

// MatchOne resolves exact label first, else unique CI substring.
// Returns error if not found or ambiguous.
func (f *File) MatchOne(query string) (*Task, error) {
	if query == "" {
		return nil, fmt.Errorf("task label required")
	}
	// exact
	for i := range f.Tasks {
		if f.Tasks[i].Label == query {
			return &f.Tasks[i], nil
		}
	}
	matches := f.FindSubstring(query)
	if len(matches) == 0 {
		return nil, fmt.Errorf("task not found: %q", query)
	}
	if len(matches) > 1 {
		var labels []string
		for _, m := range matches {
			labels = append(labels, m.Label)
		}
		return nil, fmt.Errorf("ambiguous task %q matches multiple: %s", query, strings.Join(labels, ", "))
	}
	return matches[0], nil
}
