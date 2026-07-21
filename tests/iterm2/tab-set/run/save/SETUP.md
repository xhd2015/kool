# Scenario

**Feature**: tab-set run --save (ad-hoc tabs → JSON; never iTerm)

```
# create
Caller run <name> --tab … --save [--force]
  -> save planner -> write KOOL_ITERM2_TAB_SET_DIR/<name>.json
  -> RunTabSet not called

# dry-run save
--save --dry-run -> plan/diff only; no disk write

# errors
--save without --tab | --save -n | non-TTY overwrite without --force
  -> Error exit ≠ 0; no write (or no change)
```

## Steps

1. Inherit Subcommand `run`.
2. Leaves set `Save=true`, `Tabs` (except without-tab leaf), optional Force /
   DryRun / WindowName / NewWindow.
3. Create leaves start from empty ConfigDir; overwrite leaves pre-write JSON.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markRunTree()
	markTabSetRunTree()
	markTabSetTree()
	markRootTree()
	req.Save = true
	return nil
}

// markTabSetRunSaveTree keeps hierarchical child packages importing this package live.
func markTabSetRunSaveTree() {}

// markRunSaveTree keeps hierarchical child packages importing this package live.
func markRunSaveTree() {}
```
