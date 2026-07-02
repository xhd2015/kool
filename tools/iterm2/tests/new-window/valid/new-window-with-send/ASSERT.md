## Expected

- ExitCode is 0
- Script contains `create window with default profile`
- Script contains `write text "echo hi"` (follow-up command)
- Script does NOT contain `set matchingWindow`

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
    if !strings.Contains(resp.ScriptText, "create window with default profile") {
        t.Fatal("should create a new window")
    }
    if !strings.Contains(resp.ScriptText, `write text "echo hi"`) {
        t.Fatalf("should include follow-up command 'echo hi', got script: %s", resp.ScriptText)
    }
    if strings.Contains(resp.ScriptText, "set matchingWindow") {
        t.Fatal("should NOT scan sessions when -n is set")
    }
}
```
