## Expected

- Exit 0; script includes path scan and a **match** branch that focuses (e.g. `select`)
  the tab/session at `targetDir`, not only a blind cd in the front terminal.
- Match branch has no `write text` and no `create tab`.

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
	match := reuseScriptMatchBranch(s)
	if match == "" {
		t.Fatalf("missing match branch: %q", s)
	}
	if !strings.Contains(match, "select") {
		t.Fatalf("match branch must focus (select): %q", match)
	}
	if !strings.Contains(match, "matchingSession") && !strings.Contains(match, "matchingTab") {
		t.Fatalf("match branch must focus tab/session (matchingTab or matchingSession): %q", match)
	}
	matchBranchMustNotContain(t, s, "write text")
	matchBranchMustNotContain(t, s, "create tab with default profile")
	if strings.Contains(s, "current session of current tab of current window") {
		t.Fatal("reuse -r must not use legacy current-session cd target")
	}
}
```