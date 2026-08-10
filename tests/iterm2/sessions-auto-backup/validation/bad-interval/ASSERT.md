## Expected

- Non-zero exit
- Error: mentions interval / duration / invalid (wording flexible)
- never-written.json not created

## Errors

- Invalid --interval rejected before loop

## Exit Code

- ≠ 0

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
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for bad interval; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	msg := resp.Stderr + "\n" + resp.Stdout
	if !strings.Contains(msg, "Error:") && !strings.Contains(strings.ToLower(msg), "error") {
		t.Fatalf("expected Error: for bad interval; stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}
	low := strings.ToLower(msg)
	if !strings.Contains(low, "interval") &&
		!strings.Contains(low, "duration") &&
		!strings.Contains(low, "invalid") &&
		!strings.Contains(low, "parse") {
		t.Fatalf("error should mention interval/duration; msg=%q", msg)
	}
	p := filepath.Join(req.WorkingDir, "never-written.json")
	if _, e := os.Stat(p); !os.IsNotExist(e) {
		t.Fatalf("must not write %s on validation failure", p)
	}
}
```
