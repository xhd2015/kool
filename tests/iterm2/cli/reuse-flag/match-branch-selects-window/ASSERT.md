## Expected

- Exit 0; match branch selects `matchingWindow` at the application level (bring window
  to front) before or together with tab/session focus. Tab-only `select matchingTab`
  inside `tell matchingWindow` does not activate a background window.

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
	match := reuseScriptMatchBranch(s)
	if match == "" {
		t.Fatalf("missing match branch: %q", s)
	}
	if !strings.Contains(s, "select matchingWindow") {
		t.Fatalf("reuse match branch must select matchingWindow at app level to bring window front: %q", match)
	}
	if !strings.Contains(match, "select matchingTab") || !strings.Contains(match, "matchingSession") {
		t.Fatalf("reuse match branch must still focus tab and session: %q", match)
	}
	matchBranchMustNotContain(t, s, "write text")
	matchBranchMustNotContain(t, s, "create tab with default profile")
}
```