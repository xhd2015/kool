## Expected

- ExitCode is 0
- Stdout is empty
- Stderr is empty
- AppleScript is non-empty
- Script contains `set matchingWindow` (scan logic, ModeSmart)
- Script contains `create tab with default profile` (reuse branch)
- Script contains `create window with default profile` (new window branch)
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
    if resp.Stderr != "" {
        t.Fatalf("expected empty stderr, got %q", resp.Stderr)
    }
    if resp.ScriptText == "" {
        t.Fatal("expected non-empty AppleScript")
    }
    if !strings.Contains(resp.ScriptText, "set matchingWindow") {
        t.Fatal("ModeSmart script should contain session scanning logic")
    }
    if !strings.Contains(resp.ScriptText, "create tab with default profile") {
        t.Fatal("ModeSmart script should contain tab creation branch")
    }
    if !strings.Contains(resp.ScriptText, "create window with default profile") {
        t.Fatal("ModeSmart script should contain new window branch")
    }
    if !strings.Contains(resp.ScriptText, "user.koolTargetDir") {
        t.Fatal("ModeSmart script should set koolTargetDir")
    }
}
```
