## Expected

- Exit 0
- Stdout Would save + grok session id + mark
- plan.json not created

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
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
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
	p := filepath.Join(req.WorkingDir, "plan.json")
	if _, e := os.Stat(p); !os.IsNotExist(e) {
		t.Fatalf("dry-run must not write %s", p)
	}
}
```
