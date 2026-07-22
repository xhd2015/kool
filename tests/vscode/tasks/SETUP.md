# Scenario

**Feature**: kool vscode tasks — discover tasks.json, list/find/show, dry-run plan, local/iterm2 run backends (live RunTabSet via mock env)

```
# discovery
Caller --dir | cwd
  -> walk-up to <workspace>/.vscode/tasks.json
  -> JSONC loader -> Task model

# list / find / show
Task model -> matcher / table / JSON
  -> stdout success | stderr Error + exit 1

# run --dry-run
label match -> dependsOn expand -> var expand
  -> plan on stdout; backend=iterm2 adds tab plan (offline)
  -> never calls RunTabSet

# run (execute)
label match -> plan expand
  -> backend auto|local|iterm2
  -> local: spawn leaves (sequential multi)
  -> iterm2 live: RunTabSet(ephemeral TabSetSpec, mode)
       | mock env KOOL_VSCODE_TASKS_ITERM2_MOCK* -> JSON call record (CI)
       | error -> fail closed (no local fallback)
```

## Preconditions

- Product command `kool vscode tasks` is Classic TDD target — may be missing
  entirely (leaves RED until implementer wires handler + CLI routing).
- P1 leaves (list/find/show/dry-run/match/help/validation) must remain valid.
- P2: real `run` without `--dry-run` executes; no P1 "requires dry-run" seam.
- Live iterm2: product must honor mock env (see helpers) instead of stubbing
  with `live iterm2 run not configured` + local fallback.
- Workspace discovery: start at `--dir` or process cwd; find
  `<root>/.vscode/tasks.json`.
- JSONC: `//` comments and trailing commas allowed.
- Fixtures live under leaf temp dirs via helpers (no network; no live iTerm in CI).
- Phase default `cli`: requires built `kool` binary (`go build -o kool .`).

## Steps

1. Root Setup creates temp `WorkingDir` as empty workspace sandbox.
2. Leaves write `.vscode/tasks.json` (or nested layouts) via helpers.
3. Run shells `kool vscode tasks …` with WorkingDir / `--dir` / Env / Backend as set.
4. Live iterm2 leaves call `enableITerm2Mock` / `enableITerm2MockErr` before Run.

## Context

- Nested doctest root under `tests/vscode/tasks/` — does not inherit
  `tests/vscode/open` Request/Run.
- Shared multi-task fixture includes shell leaf, process leaf, composite, and
  JSONC comments for realistic parsing.
- P2 helpers: `echoLeafJSONC`, `echoCompositeJSONC`, `failLeafJSONC` for offline
  local exec fixtures (`echo` / `false` only).
- Live mock helpers: `enableITerm2Mock`, `enableITerm2MockErr`, `readITerm2MockCall`,
  `mockMode`, `mockTabCount`, `mockWindowName`.
- Helpers: `writeTasksJSON`, `writeMultiTaskFixture`, `tasksJSONPath`,
  `combinedOut`.

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// multiTaskJSONC: shell + process + composite + comments + trailing commas.
// Labels sorted alphabetically for list assertions:
//   Build All (composite), Compile (shell), Serve (process background)
const multiTaskJSONC = `{
  "version": "2.0.0",
  // sample workspace tasks for doctest
  "tasks": [
    {
      "label": "Build All",
      "dependsOn": [
        "Compile",
        "Serve",
      ],
      "group": {
        "kind": "build",
        "isDefault": true,
      },
    },
    {
      "label": "Compile",
      "type": "shell",
      "command": "go build -o bin/app ./cmd",
      "options": {
        "cwd": "${workspaceFolder}",
      },
    },
    {
      "label": "Serve",
      "type": "process",
      "command": "bin/app",
      "args": ["--port", "8080"],
      "isBackground": true,
      "options": {
        "cwd": "${workspaceFolder}",
      },
    },
  ],
}
`

// expandFixture: leaf with workspaceFolder + env var for dry-run / real-run expansion.
const expandTaskJSONC = `{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Echo Root",
      "type": "shell",
      "command": "echo ${workspaceFolder} ${workspaceFolderBasename} ${env:KOOL_TASKS_TEST_TOKEN}",
      "options": {
        "cwd": "${workspaceFolder}"
      }
    }
  ]
}
`

// unresolvedVarJSONC: unknown ${…} token.
const unresolvedVarJSONC = `{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Bad Var",
      "type": "shell",
      "command": "echo ${workspaceFolder} ${unknownToken}"
    }
  ]
}
`

// cycleJSONC: A -> B -> A
const cycleJSONC = `{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Cycle A",
      "dependsOn": ["Cycle B"]
    },
    {
      "label": "Cycle B",
      "dependsOn": ["Cycle A"]
    }
  ]
}
`

// missingDepJSONC: depends on absent label
const missingDepJSONC = `{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Root",
      "dependsOn": ["No Such Task"]
    }
  ]
}
`

// ambiguous pair for show/find/run match tests
const ambiguousPairJSONC = `{
  "version": "2.0.0",
  "tasks": [
    {"label": "Alpha One", "type": "shell", "command": "echo a1"},
    {"label": "Alpha Two", "type": "shell", "command": "echo a2"},
    {"label": "Beta", "type": "shell", "command": "echo b"}
  ]
}
`

// P2 offline exec fixtures (echo / false only — no network, no iTerm).
const echoLeafJSONC = `{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Say Hello",
      "type": "shell",
      "command": "echo KOOL_TASKS_P2_HELLO"
    }
  ]
}
`

const echoCompositeJSONC = `{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Step One",
      "type": "shell",
      "command": "echo KOOL_TASKS_P2_STEP_ONE"
    },
    {
      "label": "Step Two",
      "type": "shell",
      "command": "echo KOOL_TASKS_P2_STEP_TWO"
    },
    {
      "label": "Both Steps",
      "dependsOn": ["Step One", "Step Two"]
    }
  ]
}
`

const failLeafJSONC = `{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Fail Fast",
      "type": "shell",
      "command": "false"
    }
  ]
}
`

// Mock call JSON shape product should write under KOOL_VSCODE_TASKS_ITERM2_MOCK_OUT.
// Implementer may add fields; helpers read flexibly.
//
//	{
//	  "mode": "smart" | "new-window" | "no-new-window",
//	  "spec": {
//	    "name": "vscode-tasks-…",
//	    "windowName": "Both Steps",
//	    "tabs": [
//	      {"id": "…", "name": "Step One", "command": "echo …", "cwd": "…"}
//	    ]
//	  }
//	}
type iterm2MockCall struct {
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

func Setup(t *testing.T, req *Request) error {
	if req.WorkingDir == "" {
		req.WorkingDir = t.TempDir()
	}
	if req.Phase == "" {
		req.Phase = "cli"
	}
	return nil
}

func writeTasksJSON(t *testing.T, workspace, content string) {
	t.Helper()
	dir := filepath.Join(workspace, ".vscode")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "tasks.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeMultiTaskFixture(t *testing.T, workspace string) {
	t.Helper()
	writeTasksJSON(t, workspace, multiTaskJSONC)
}

func writeEmptyTasks(t *testing.T, workspace string) {
	t.Helper()
	writeTasksJSON(t, workspace, `{
  "version": "2.0.0",
  "tasks": []
}
`)
}

func tasksJSONPath(workspace string) string {
	return filepath.Join(workspace, ".vscode", "tasks.json")
}

func combinedOut(resp *Response) string {
	return resp.Stdout + resp.Stderr
}

// workspaceBasename returns filepath.Base(workspace) for plan assertions.
func workspaceBasename(workspace string) string {
	return filepath.Base(workspace)
}

// enableITerm2Mock installs CI mock env: product must write RunTabSet call JSON
// to the returned path and must not open live iTerm. Does not force error.
func enableITerm2Mock(t *testing.T, req *Request) string {
	t.Helper()
	if req.WorkingDir == "" {
		req.WorkingDir = t.TempDir()
	}
	out := filepath.Join(req.WorkingDir, "iterm2-mock-call.json")
	// remove stale file so "file exists" means product wrote it this run
	_ = os.Remove(out)
	req.ITerm2MockOut = out
	req.Env = append(req.Env,
		"KOOL_VSCODE_TASKS_ITERM2_MOCK=1",
		"KOOL_VSCODE_TASKS_ITERM2_MOCK_OUT="+out,
	)
	return out
}

// enableITerm2MockErr is enableITerm2Mock plus MOCK_ERR=1 (product returns error
// after recording the call when possible).
func enableITerm2MockErr(t *testing.T, req *Request) string {
	t.Helper()
	out := enableITerm2Mock(t, req)
	req.Env = append(req.Env, "KOOL_VSCODE_TASKS_ITERM2_MOCK_ERR=1")
	return out
}

func readITerm2MockCall(t *testing.T, path string) iterm2MockCall {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read mock call JSON %s: %v (product must write RunTabSet record when KOOL_VSCODE_TASKS_ITERM2_MOCK=1)", path, err)
	}
	var call iterm2MockCall
	if err := json.Unmarshal(data, &call); err != nil {
		t.Fatalf("parse mock call JSON: %v\nbody:\n%s", err, data)
	}
	return call
}

func mockFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func normalizeMode(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

// assertNoLiveStubWarning fails if product still uses the pre-wire local stub.
func assertNoLiveStubWarning(t *testing.T, out string) {
	t.Helper()
	lower := strings.ToLower(out)
	if strings.Contains(lower, "live iterm2 run not configured") {
		t.Fatalf("product still stubs live iterm2 with local fallback; out:\n%s", out)
	}
}
```
