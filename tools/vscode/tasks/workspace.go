package tasks

import (
	"fmt"
	"os"
	"path/filepath"
)

// Workspace holds the resolved root and tasks.json path.
type Workspace struct {
	Root     string // directory containing .vscode
	TasksPath string
}

// FindWorkspace walks up from startDir (or cwd) until .vscode/tasks.json exists.
func FindWorkspace(startDir string) (*Workspace, error) {
	dir := startDir
	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		dir = cwd
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	dir = abs
	for {
		candidate := filepath.Join(dir, ".vscode", "tasks.json")
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return &Workspace{Root: dir, TasksPath: candidate}, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return nil, fmt.Errorf("tasks.json not found: no .vscode/tasks.json in %s or parent directories", startDir)
}

// LoadTasks reads and parses tasks.json from the workspace.
func LoadTasks(ws *Workspace) (*File, error) {
	data, err := os.ReadFile(ws.TasksPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read tasks.json: %w", err)
	}
	f, err := ParseTasksJSON(data)
	if err != nil {
		return nil, err
	}
	return f, nil
}
