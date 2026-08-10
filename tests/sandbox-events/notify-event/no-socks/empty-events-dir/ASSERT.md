## Expected

- Exit 0 (no subscribers is not an error).
- Stderr contains a warning about no sockets / no subscribers / empty events
  (case-insensitive match on at least one of: warn, no, sock, subscriber, empty).
- Must not panic or return unrecognized-command as the only outcome once
  implemented (pre-impl RED: non-zero unrecognized is also failure here).

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("empty events should exit 0; exit=%d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	low := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	// Prefer stderr warning; allow soft wording variants.
	hasWarn := strings.Contains(low, "warn") ||
		strings.Contains(low, "no sock") ||
		strings.Contains(low, "no subscriber") ||
		strings.Contains(low, "no listener") ||
		strings.Contains(low, "empty") ||
		(strings.Contains(low, "no") && strings.Contains(low, "event"))
	if !hasWarn {
		t.Fatalf("expected warning about no socks/subscribers; stderr=%q stdout=%q",
			resp.Stderr, resp.Stdout)
	}
}
```
