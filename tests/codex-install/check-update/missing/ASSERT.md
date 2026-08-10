## Expected

- Exit **0** (success path even when missing).
- Stdout reports missing / not installed / not found (case-insensitive).
- `ShellCalls` empty.
- `FetchLatestCalls == 0` (missing path must not require latest).

## Side Effects

- No `RunShell` calls.
- No latest fetch.

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
	if resp.FetchLatestCalls != 0 {
		t.Fatalf("FetchLatestCalls = %d, want 0 when bin missing", resp.FetchLatestCalls)
	}
	low := strings.ToLower(resp.Stdout)
	missing := strings.Contains(low, "missing") ||
		strings.Contains(low, "not installed") ||
		strings.Contains(low, "not found") ||
		strings.Contains(low, "absent")
	if !missing {
		t.Fatalf("check-update missing should report missing/not installed; got:\n%s", resp.Stdout)
	}
}
```
