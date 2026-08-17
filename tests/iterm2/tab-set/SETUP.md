# Scenario

**Feature**: kool iterm2 tab-set CLI with injectable config directory

```
# list / show / dry-run (no iTerm)
KOOL_ITERM2_TAB_SET_DIR=<tmp>
  -> kool iterm2 tab-set <cmd> …
  -> load <name>.json -> stdout / validation errors

# run --dry-run (config mode)
config tabs -> plan printed (no_submit tabs marked); RunTabSet Exec not required

# run --tab … --dry-run (ad-hoc; no config read)
Caller --tab specs (props: id,name,cwd,no_submit) -> tab parser -> plan

# run --tab … --save [--force|--dry-run]
Caller tabs -> save planner -> JSON write (incl. no_submit) or plan; never RunTabSet

# update <name> [--tab-id …] [--rm|--command|--no-submit|…]
# Caller patches config JSON only; never RunTabSet
config file -> update mutator -> validate -> write (or dry-run plan)
```

## Preconditions

- Package `github.com/xhd2015/kool/tools/iterm2` exports `RunForTest` (exists).
- Config dir: env `KOOL_ITERM2_TAB_SET_DIR` (default `~/.config/iterm2/tab-set`).
- Version-1 JSON schema with validation (version, tabs, ids, commands).
- Optional per-tab `no_submit` bool.
- `update` subcommand (Classic TDD this cycle — new leaves RED until implementer).

## Steps

1. Root Setup creates temp `WorkingDir` and empty `ConfigDir`.
2. Leaves write JSON fixtures under `ConfigDir` (config mode) and/or set
   `Tabs` / `Save` / `Force` / `WindowName` (ad-hoc/save) or update flags
   (`TabID`, `Rm`, `Command`, `UpdateNoSubmit`, …).
3. Run invokes in-process `RunForTestEnv` with `TestRun.TabSetDir` (no process env).

## Context

- Nested doctest root — does not inherit open-dir `Request` from `../DOCTEST.md`.
- Prefer Phase=`handler` (in-process); no live iTerm for these leaves.
- Fixture helper writes `bots.json` matching the locked schema.
- `configPath(configDir, name)` helper for save/update leaf file checks.
- Handler RunForTest is non-TTY (bytes.Buffer) — non-TTY overwrite/rm path is
  the CI-friendly assert for confirm rules (`--force` required for success).

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
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

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
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
```
