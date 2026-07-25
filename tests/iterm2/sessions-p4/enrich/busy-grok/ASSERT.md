## Expected

- Exit 0.
- Stdout contains fixture agent session id
  `019fabcdef-1234-5678-9abc-def012345678`.
- Stdout contains runner kind `grok`.
- Stdout contains at least one Unicode tree connector used by FormatTree:
  `└──` and/or `├──` (busy-grok fixture is a sole-child chain → typically `└──`).
- Stdout still shows busy pane context (e.g. `busy` or busy session id fragment).

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
	if !strings.Contains(out, "grok") {
		t.Fatalf("stdout missing grok kind/marker:\n%s", out)
	}
	hasConnector := strings.Contains(out, "└──") || strings.Contains(out, "├──")
	if !hasConnector {
		t.Fatalf("stdout missing Unicode tree connectors (└──/├──):\n%s", out)
	}
	// Busy pane still present (process enrich + hierarchy).
	if !strings.Contains(out, "busy") && !strings.Contains(strings.ToLower(out), "bbbbbbbb") {
		t.Fatalf("stdout missing busy pane marker:\n%s", out)
	}
}
```
