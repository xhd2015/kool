## Expected

- ExitCode is 0
- Script same as -r flag: contains `select matchingWindow`, no `create tab`
- Behavior identical to reuse-r test

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
        t.Fatal("--reuse-window should produce ModeReuseCurrent script with select matchingWindow")
    }
    if strings.Contains(resp.ScriptText, "create tab with default profile") {
        t.Fatal("--reuse-window script should NOT contain tab creation")
    }
}
```
