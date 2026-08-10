## Expected

- Exit **0** (success path even though outdated).
- Stdout reports update available / needs update / outdated (case-insensitive).
- Stdout includes versions `0.1.0` and `0.2.0`.
- `ShellCalls` empty (status only; no install/update).

## Side Effects

- No `RunShell` calls.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	assertExitZero(t, resp)
	assertNoShell(t, resp)
	assert.Output(t, resp.Stdout, `<contains>
0.1.0
0.2.0
</contains>
`)
	low := strings.ToLower(resp.Stdout)
	available := strings.Contains(low, "update available") ||
		strings.Contains(low, "needs update") ||
		strings.Contains(low, "outdated") ||
		strings.Contains(low, "update:") ||
		(strings.Contains(low, "update") && (strings.Contains(low, "available") || strings.Contains(low, "needed")))
	if !available {
		t.Fatalf("check-update outdated should report update available; got:\n%s", resp.Stdout)
	}
}
```
