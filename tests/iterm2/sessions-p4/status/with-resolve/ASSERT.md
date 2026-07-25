## Expected

- Exit 0.
- Stdout contains fixture agent session id
  `019fabcdef-1234-5678-9abc-def012345678` (runner session).
- Stdout contains `grok`.
- Stdout references the iTerm session under status (full or short busy id
  `BBBBBBBB…` / name `busy-sess`).
- Tree connector `└──` or `├──` present (status enrich parity).

## Errors

- None.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	if !strings.Contains(out, fixtureGrokSessionID) {
		t.Fatalf("status missing runner session id %q:\n%s", fixtureGrokSessionID, out)
	}
	if !strings.Contains(out, "grok") {
		t.Fatalf("status missing grok:\n%s", out)
	}
	// iTerm session identity still present.
	if !strings.Contains(out, fixtureITermBusyID) &&
		!strings.Contains(strings.ToLower(out), "bbbbbbbb") &&
		!strings.Contains(out, "busy-sess") {
		t.Fatalf("status missing iTerm busy session marker:\n%s", out)
	}
	if !strings.Contains(out, "└──") && !strings.Contains(out, "├──") {
		t.Fatalf("status missing tree connector (parity with snapshot enrich):\n%s", out)
	}
}
```
