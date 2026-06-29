# Scenario

**Feature**: successful CLI runs capture AppleScript via fake osascript

```
kool iterm2 <dir> [--send ...] -> exit 0 -> script file contains cd + follow-ups
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "cli"
	req.InstalledEnv = "1"
	return nil
}
```