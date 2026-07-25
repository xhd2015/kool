## Expected

- Exit 0.
- Stdout contains `idle` and `busy` (process enrich applied before/at emit).
- Stdout contains session id short forms or full fixture markers (`AAAAAAAA` or
  idle-sess / busy-sess) so enrich is tied to real sessions.
- Progressive stream still holds: `SawW1BeforeLastListTabs` true when probe
  active.

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
	if !strings.Contains(out, "idle") {
		t.Fatalf("stdout missing idle enrich:\n%s", out)
	}
	if !strings.Contains(out, "busy") {
		t.Fatalf("stdout missing busy enrich:\n%s", out)
	}
	// Tie to fixture sessions (command name or id fragment).
	if !strings.Contains(out, "idle-sess") && !strings.Contains(out, "AAAAAAAA") && !strings.Contains(out, "aaaa") {
		// shortID of AAAAAAAA-… is often first 8 hex without dashes → aaaaaaaa
		if !strings.Contains(strings.ToLower(out), "aaaaaaaa") {
			t.Fatalf("stdout missing fixture idle session marker:\n%s", out)
		}
	}
	if resp.LastListTabsInvoked && !resp.SawW1BeforeLastListTabs {
		t.Fatalf("enrich leaf expects streaming order (W1 before last ListTabs);\n%s", out)
	}
}
```
