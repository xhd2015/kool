## Expected

- Exit 0.
- No stub warning `live iterm2 run not configured`.
- Mock JSON `mode` normalizes to `smart` (or combined output contains `mode: smart` / `mode smart`).

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
	out := combinedOut(resp)
	if resp.ExitCode != 0 {
		t.Fatalf("iterm2 mode smart exit=%d out:\n%s", resp.ExitCode, out)
	}
	assertNoLiveStubWarning(t, out)

	if !mockFileExists(req.ITerm2MockOut) {
		t.Fatalf("mock call missing at %s", req.ITerm2MockOut)
	}
	call := readITerm2MockCall(t, req.ITerm2MockOut)
	mode := normalizeMode(call.Mode)
	if mode == "smart" {
		return
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "mode: smart") || strings.Contains(lower, "mode smart") ||
		strings.Contains(lower, `"mode":"smart"`) || strings.Contains(lower, "run mode: smart") {
		return
	}
	t.Fatalf("want mode smart, got mode=%q call=%+v out:\n%s", call.Mode, call, out)
}
```
