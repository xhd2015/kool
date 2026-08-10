## Expected

- Exit 0.
- `ShellCalls` is exactly `[install.InstallCmd]`.
- `FetchLatestCalls == 0` (latest only when bin present).

## Side Effects

- Exactly one shell mutator: library `InstallCmd`.

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
	assertShellCalls(t, resp.ShellCalls, codexinstall.InstallCmd)
	if resp.FetchLatestCalls != 0 {
		t.Fatalf("FetchLatestCalls = %d, want 0 when bin missing", resp.FetchLatestCalls)
	}
}
```
