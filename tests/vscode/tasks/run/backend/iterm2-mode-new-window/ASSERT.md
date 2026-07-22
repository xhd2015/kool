## Expected

- Exit 0.
- No stub warning.
- Mock `mode` is `new-window` (accept `new_window` / `newwindow` via normalize).

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
		t.Fatalf("iterm2 -n mode exit=%d out:\n%s", resp.ExitCode, out)
	}
	assertNoLiveStubWarning(t, out)

	if !mockFileExists(req.ITerm2MockOut) {
		t.Fatalf("mock call missing at %s", req.ITerm2MockOut)
	}
	call := readITerm2MockCall(t, req.ITerm2MockOut)
	mode := normalizeMode(call.Mode)
	// normalizeMode maps spaces/underscores to hyphens
	ok := mode == "new-window" || mode == "newwindow" || strings.Contains(mode, "new-window")
	if !ok {
		lower := strings.ToLower(out)
		ok = strings.Contains(lower, "new-window") || strings.Contains(lower, "mode: new")
	}
	if !ok {
		t.Fatalf("want mode new-window, got mode=%q call=%+v out:\n%s", call.Mode, call, out)
	}
}
```
