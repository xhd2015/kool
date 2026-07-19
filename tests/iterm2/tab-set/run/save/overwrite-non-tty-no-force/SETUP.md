# Scenario

**Feature**: overwrite without --force on non-TTY errors and does not write

```
existing bots.json + run bots --tab … --save  (handler non-TTY, no --force)
  -> Error exit ≠ 0; bots.json content unchanged
```

## Steps

1. Write bots fixture; record that content is the baseline.
2. Save with different tabs; Force=false; no DryRun.
3. RunForTest is non-interactive (no TTY) — locked rule requires --force.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeBotsConfig(t, req.ConfigDir)
	req.SetName = "bots"
	req.Save = true
	req.Force = false
	req.Tabs = []string{
		"[id=z] echo replaced-all",
	}
	return nil
}
```
