## Expected

- Exit 0
- Stdout contains ANSI escape (`\x1b[`)
- Plan still has Would save (possibly inside color spans) and critical content
- No file written

## Errors

- None

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
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("--color must emit ANSI escapes; stdout:\n%s", out)
	}
	// Green verb and/or bold window label (token map D6/D7).
	hasGreenOrBold := strings.Contains(out, "\x1b[32m") || // green
		strings.Contains(out, "\x1b[1m") || // bold
		strings.Contains(out, "\x1b[01m")
	if !hasGreenOrBold {
		t.Fatalf("expected green and/or bold ANSI (Would save / W{n}); stdout:\n%q", out)
	}
	if !strings.Contains(out, "Would save") {
		t.Fatalf("missing Would save (may be colored):\n%s", out)
	}
	p := filepath.Join(req.WorkingDir, "color-plan.json")
	if _, e := os.Stat(p); !os.IsNotExist(e) {
		t.Fatalf("dry-run must not write %s", p)
	}
}
```
