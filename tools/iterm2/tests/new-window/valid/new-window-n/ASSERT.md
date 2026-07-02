## Expected

- ExitCode is 0
- Script does NOT contain `set matchingWindow` (no scanning)
- Script contains `create window with default profile`
- Script does NOT contain `create tab with default profile`
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
    if strings.Contains(resp.ScriptText, "set matchingWindow") {
        t.Fatal("ModeForceNew script should NOT contain session scanning")
    }
    if !strings.Contains(resp.ScriptText, "create window with default profile") {
        t.Fatal("ModeForceNew script should create a new window")
    }
    if strings.Contains(resp.ScriptText, "create tab with default profile") {
        t.Fatal("ModeForceNew script should NOT contain tab creation")
    }
    if !strings.Contains(resp.ScriptText, "user.koolTargetDir") {
        t.Fatal("ModeForceNew script should set koolTargetDir")
    }
}
```
