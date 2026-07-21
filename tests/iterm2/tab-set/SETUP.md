# Scenario

**Feature**: kool iterm2 tab-set CLI with injectable config directory

```
# list / show / dry-run (no iTerm)
KOOL_ITERM2_TAB_SET_DIR=<tmp>
  -> kool iterm2 tab-set <cmd> …
  -> load <name>.json -> stdout / validation errors

# run --dry-run (config mode)
config tabs -> plan printed; RunTabSet Exec not required

# run --tab … --dry-run (ad-hoc; no config read)
Caller --tab specs -> tab parser -> plan; config file not required

# run --tab … --save [--force|--dry-run]
Caller tabs -> save planner -> JSON write or plan; never RunTabSet
```

## Preconditions

- Package `github.com/xhd2015/kool/tools/iterm2` exports `RunForTest` (exists).
- Config dir: env `KOOL_ITERM2_TAB_SET_DIR` (default `~/.config/iterm2/tab-set`).
- Version-1 JSON schema with validation (version, tabs, ids, commands).
- Ad-hoc/save flags (`--tab`, `--save`, `--force`, `--window-name`) are Classic
  TDD targets for this cycle — product may not wire them yet (new leaves RED).

## Steps

1. Root Setup creates temp `WorkingDir` and empty `ConfigDir`.
2. Leaves write JSON fixtures under `ConfigDir` (config mode) and/or set
   `Tabs` / `Save` / `Force` / `WindowName` (ad-hoc/save mode).
3. Run invokes in-process `RunForTest` with `KOOL_ITERM2_TAB_SET_DIR` set.

## Context

- Nested doctest root — does not inherit open-dir `Request` from `../DOCTEST.md`.
- Prefer Phase=`handler` (in-process); no live iTerm for these leaves.
- Fixture helper writes `bots.json` matching the locked schema.
- `configPath(configDir, name)` helper for save-leaf file checks.
- Handler RunForTest is non-TTY (bytes.Buffer) — non-TTY overwrite path is
  the CI-friendly assert for confirm rules.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

const botsJSON = `{
  "version": 1,
  "window_name": "local-bots",
  "tabs": [
    {"id": "a", "name": "a", "command": "echo a"},
    {"id": "b", "name": "b", "command": "echo b", "cwd": "/tmp"}
  ]
}
`

func Setup(t *testing.T, req *Request) error {
	markRootTree()
	if req.WorkingDir == "" {
		req.WorkingDir = t.TempDir()
	}
	if req.ConfigDir == "" {
		req.ConfigDir = filepath.Join(req.WorkingDir, "tab-set-config")
		if err := os.MkdirAll(req.ConfigDir, 0755); err != nil {
			return err
		}
	}
	if req.Phase == "" {
		req.Phase = "handler"
	}
	if req.GoOS == "" {
		req.GoOS = "darwin"
	}
	return nil
}

func writeConfigFile(t *testing.T, configDir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(configDir, name+".json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeBotsConfig(t *testing.T, configDir string) {
	t.Helper()
	writeConfigFile(t, configDir, "bots", botsJSON)
}

func configPath(configDir, name string) string {
	return filepath.Join(configDir, name+".json")
}

func combinedOut(resp *Response) string {
	return resp.Stdout + resp.Stderr
}

// markTabSetTree keeps hierarchical child packages importing this package live.
func markTabSetTree() {}

// markRootTree keeps hierarchical child packages importing this package live.
func markRootTree() {}
```
