## Expected

- Exit 0.
- No stub warning.
- Mock `mode` is `no-new-window`.

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
		t.Fatalf("iterm2 --no-new-window mode exit=%d out:\n%s", resp.ExitCode, out)
	}
	assertNoLiveStubWarning(t, out)

	if !mockFileExists(req.ITerm2MockOut) {
		t.Fatalf("mock call missing at %s", req.ITerm2MockOut)
	}
	call := readITerm2MockCall(t, req.ITerm2MockOut)
	mode := normalizeMode(call.Mode)
	ok := mode == "no-new-window" || mode == "nonewwindow" || strings.Contains(mode, "no-new-window")
	if !ok {
		lower := strings.ToLower(out)
		ok = strings.Contains(lower, "no-new-window")
	}
	if !ok {
		t.Fatalf("want mode no-new-window, got mode=%q call=%+v out:\n%s", call.Mode, call, out)
	}
}
```
