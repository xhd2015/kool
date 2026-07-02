## Expected

- ExitCode is 0
- Script same as -n flag: no scanning, always new window
- Behavior identical to new-window-n test

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
        t.Fatal("--new-window should NOT contain session scanning")
    }
    if !strings.Contains(resp.ScriptText, "create window with default profile") {
        t.Fatal("--new-window script should create a new window")
    }
}
```
