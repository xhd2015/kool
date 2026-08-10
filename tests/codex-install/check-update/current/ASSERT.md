## Expected

- Exit 0.
- Stdout reports up to date / current / no update needed (case-insensitive).
- Must **not** claim update available / outdated as the primary status.
- `ShellCalls` empty.

## Side Effects

- No `RunShell` calls.

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
	assertNoError(t, err)
	assertExitZero(t, resp)
	assertNoShell(t, resp)
	low := strings.ToLower(resp.Stdout)
	ok := strings.Contains(low, "up to date") ||
		strings.Contains(low, "up-to-date") ||
		strings.Contains(low, "current") ||
		strings.Contains(low, "no update") ||
		strings.Contains(low, "already")
	if !ok {
		t.Fatalf("check-update current should report up to date; got:\n%s", resp.Stdout)
	}
	if strings.Contains(low, "update available") || strings.Contains(low, "outdated") {
		t.Fatalf("check-update current must not claim update available; got:\n%s", resp.Stdout)
	}
}
```
