## Expected

- ExitCode is 1, stderr mentions conflict, no script

```go
import (
    "strings"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    if err != nil {
        t.Fatal(err)
    }
    if resp.ExitCode != 1 {
        t.Fatalf("expected exit code 1, got %d", resp.ExitCode)
    }
    if !strings.Contains(resp.Stderr, "cannot specify both") && !strings.Contains(resp.Stderr, "mutually exclusive") {
        t.Fatalf("stderr should mention conflict, got: %q", resp.Stderr)
    }
    if resp.ScriptText != "" {
        t.Fatal("conflict should not generate AppleScript")
    }
}
```
