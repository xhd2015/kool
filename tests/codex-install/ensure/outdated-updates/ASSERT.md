## Expected

- Exit 0.
- `ShellCalls` is exactly `[install.UpdateCmd]`.
- `FetchLatestCalls >= 1`.

## Side Effects

- Exactly one shell mutator: library `UpdateCmd` (`codex update`).

## Exit Code

- 0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	codexinstall "github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/install"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	assertExitZero(t, resp)
	assertShellCalls(t, resp.ShellCalls, codexinstall.UpdateCmd)
	if resp.FetchLatestCalls < 1 {
		t.Fatalf("FetchLatestCalls = %d, want >= 1", resp.FetchLatestCalls)
	}
}
```
