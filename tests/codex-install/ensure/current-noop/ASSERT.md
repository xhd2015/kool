## Expected

- Exit 0.
- `ShellCalls` empty (noop; no install/update).
- `FetchLatestCalls >= 1`.

## Side Effects

- No shell mutators.

## Exit Code

- 0

```go
import (
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
	if resp.FetchLatestCalls < 1 {
		t.Fatalf("FetchLatestCalls = %d, want >= 1", resp.FetchLatestCalls)
	}
}
```
