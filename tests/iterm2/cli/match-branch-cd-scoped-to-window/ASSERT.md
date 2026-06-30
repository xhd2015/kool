## Expected

- Exit 0; match branch creates a tab in `matchingWindow` and scopes `cd` to that
  window's new tab — not unqualified `current session of current tab` (which targets
  the frontmost window and sends cd to an unrelated session).

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
	match := smartScriptMatchBranch(s)
	if match == "" {
		t.Fatalf("missing match branch: %q", s)
	}
	if !strings.Contains(match, "create tab with default profile") {
		t.Fatalf("match branch must create tab in matchingWindow: %q", match)
	}
	if strings.Contains(match, "tell current session of current tab\n") ||
		strings.Contains(match, "tell current session of current tab\r\n") {
		t.Fatalf("match branch must not cd via unqualified current tab (hits frontmost window): %q", match)
	}
	scoped := strings.Contains(match, "current session of current tab of matchingWindow") ||
		strings.Contains(match, "current session of newTab")
	if !scoped {
		t.Fatalf("match branch must scope cd to matchingWindow tab (e.g. newTab or current tab of matchingWindow): %q", match)
	}
}
```