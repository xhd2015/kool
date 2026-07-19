# Scenario

**Feature**: --tab … --save creates version-1 JSON; never runs iTerm

```
run mysave --tab "[id=x] echo x" --tab "[id=y] echo y" --window-name win-save --save --force
  -> write mysave.json (version 1, tabs, window_name)
  -> RunTabSet not called; exit 0
```

## Steps

1. Empty ConfigDir (no mysave.json yet).
2. Two props tabs + WindowName; Save=true; Force=true (create may not need force, but harmless).
3. Not DryRun — real write; still no iTerm because --save.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SetName = "mysave"
	req.Save = true
	req.Force = true
	req.WindowName = "win-save"
	req.Tabs = []string{
		"[id=x] echo x",
		"[id=y,cwd=/tmp] echo y",
	}
	return nil
}
```
