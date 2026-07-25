## Expected

- Exit 0.
- Stdout contains fixture agent session id
  `019fabcdef-1234-5678-9abc-def012345678`.
- Stdout contains all three Unicode FormatTree connectors: `├──`, `└──`, `│`.
- Stdout contains pids `200`, `201`, `202` and cmd substrings `agent-run` and
  `grok`.

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
		t.Fatalf("stdout missing agent session id %q:\n%s", fixtureGrokSessionID, out)
	}
	for _, glyph := range []string{"├──", "└──", "│"} {
		if !strings.Contains(out, glyph) {
			t.Fatalf("stdout missing Unicode connector %q:\n%s", glyph, out)
		}
	}
	for _, pid := range []string{"200", "201", "202"} {
		if !strings.Contains(out, pid) {
			t.Fatalf("stdout missing pid %s:\n%s", pid, out)
		}
	}
	if !strings.Contains(out, "agent-run") {
		t.Fatalf("stdout missing agent-run cmd:\n%s", out)
	}
	if !strings.Contains(out, "grok") {
		t.Fatalf("stdout missing grok cmd:\n%s", out)
	}
}
```
