## Expected

- Exit 0; captured script scans session `path` variables.
- Script does **not** target `current session of current tab of current window`.
- **Else** branch creates a new window and runs `cd`; match branch does not use `create tab`.

## Exit Code

- 0

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
		t.Fatalf("exit=%d stderr=%s", resp.ExitCode, resp.Stderr)
	}
	s := resp.CapturedScript
	if s == "" {
		t.Fatal("expected captured script")
	}
	if !scriptHasReusePathScan(s) {
		t.Fatalf("missing path scan: %q", s)
	}
	if strings.Contains(s, "current session of current tab of current window") {
		t.Fatal("reuse -r must not cd in arbitrary current session")
	}
	elseBranchMustContain(t, s, "create window with default profile")
	elseBranchMustContain(t, s, `write text ("cd " & quoted form of targetDir)`)
	matchBranchMustNotContain(t, s, "create tab with default profile")
}
```