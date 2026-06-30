# Scenario

**Feature**: successful CLI runs capture AppleScript via fake osascript

```
kool iterm2 <dir> [--send ...] -> exit 0 -> script file contains cd + follow-ups
```

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Phase = "cli"
	req.InstalledEnv = "1"
	return nil
}

func smartScriptMatchBranch(script string) string {
	const open = `if matchingWindow is not missing value then`
	start := strings.Index(script, open)
	if start < 0 {
		return ""
	}
	rest := script[start+len(open):]
	elseIdx := strings.Index(rest, "\n  else\n")
	if elseIdx < 0 {
		elseIdx = strings.Index(rest, "\n  else")
	}
	if elseIdx < 0 {
		return rest
	}
	return rest[:elseIdx]
}
```