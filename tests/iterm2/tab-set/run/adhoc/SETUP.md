# Scenario

**Feature**: tab-set run ad-hoc mode (≥1 --tab, no --save)

```
# ad-hoc dry-run: no config file read
Caller --tab "echo a" --tab "echo b" --dry-run
  -> tab parser -> dry-run plan (ids/commands)
  -> config store not consulted; RunTabSet not called

# props + defaults
--tab "  [ id = a ]  cmd" -> id=a; bare cmd -> tab-N defaults
```

## Steps

1. Inherit Subcommand `run` from parent.
2. Explicitly disable Save (ad-hoc run path, not save path).
3. Leaves set `Tabs` (≥1), usually `DryRun=true` so no iTerm is needed.
4. Do **not** write config fixtures unless a leaf needs an unrelated file present
   (ad-hoc must work with empty ConfigDir).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	markRunTree()
	markTabSetRunTree()
	markTabSetTree()
	markRootTree()
	// Ad-hoc branch: never --save (save leaves live under run/save/).
	req.Save = false
	return nil
}

// markTabSetRunAdhocTree keeps hierarchical child packages importing this package live.
func markTabSetRunAdhocTree() {}

// markRunAdhocTree keeps hierarchical child packages importing this package live.
func markRunAdhocTree() {}
```
