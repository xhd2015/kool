## Expected

- Exit 0 after one plan (no --once; proves --dry-run does not loop)
- Stdout Would save + grok session id + mark
- plan-auto.json not created

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	// Process returned: --dry-run must not hang in the interval loop without --once.
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if req.Once {
		t.Fatal("this leaf must run without --once to prove dry-run exits alone")
	}
	out := resp.Stdout
	if !strings.Contains(out, "Would save") {
		t.Fatalf("missing Would save:\n%s", out)
	}
	if !strings.Contains(out, fixtureGrokSessionID) {
		t.Fatalf("missing grok session:\n%s", out)
	}
	if !strings.Contains(out, "mark") {
		t.Fatalf("missing mark:\n%s", out)
	}
	p := filepath.Join(req.WorkingDir, "plan-auto.json")
	if _, e := os.Stat(p); !os.IsNotExist(e) {
		t.Fatalf("dry-run must not write %s", p)
	}
}
```
