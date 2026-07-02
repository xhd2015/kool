## Expected

- ExitCode is 0
- Script contains `set matchingWindow` (scan logic)
- Script contains `select matchingWindow` (focus behavior, ModeReuseCurrent)
- Script does NOT contain `create tab with default profile`
- Script contains `create window with default profile` (fallback on miss)
- Script contains `user.koolTargetDir`

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatal(err)
    }
    if resp.ExitCode != 0 {
        t.Fatalf("expected exit code 0, got %d; stderr=%q", resp.ExitCode, resp.Stderr)
    }
    if resp.ScriptText == "" {
        t.Fatal("expected non-empty AppleScript")
    }
    if !strings.Contains(resp.ScriptText, "set matchingWindow") {
        t.Fatal("ModeReuseCurrent script should contain session scanning")
    }
    if !strings.Contains(resp.ScriptText, "select matchingWindow") {
        t.Fatal("ModeReuseCurrent script should select the matching window")
    }
    if strings.Contains(resp.ScriptText, "create tab with default profile") {
        t.Fatal("ModeReuseCurrent script should NOT contain tab creation")
    }
    if !strings.Contains(resp.ScriptText, "create window with default profile") {
        t.Fatal("ModeReuseCurrent script should contain new window fallback")
    }
}
```
