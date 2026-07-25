## Expected

- Exit 0.
- `LastListTabsInvoked` true.
- **`SawW1BeforeLastListTabs` false** — buffered path does not emit during
  collection.
- Final stdout still contains `W1`, `W2`, and sessions summary footer.

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
	if !resp.LastListTabsInvoked {
		t.Fatal("expected ListTabs for last window under --no-stream collect path")
	}
	if resp.SawW1BeforeLastListTabs {
		t.Fatalf("--no-stream must not emit W1 before last ListTabs; progressive leak?\n%s", resp.Stdout)
	}
	out := resp.Stdout
	for _, want := range []string{"W1", "W2", "sessions"} {
		if !strings.Contains(out, want) {
			t.Fatalf("final buffered stdout missing %q:\n%s", want, out)
		}
	}
}
```
