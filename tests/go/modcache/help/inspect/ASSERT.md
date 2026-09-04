## Expected

- Exit 0.
- Stdout mentions `inspect` and `--modcache`.
- Trailing newline.

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
		t.Fatalf("exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("help stdout must end with newline; got %q", resp.Stdout)
	}
	low := strings.ToLower(resp.Stdout)
	if !strings.Contains(low, "inspect") {
		t.Fatalf("inspect help should mention inspect; got:\n%s", resp.Stdout)
	}
	if !strings.Contains(low, "--modcache") {
		t.Fatalf("inspect help should mention --modcache; got:\n%s", resp.Stdout)
	}
	if !strings.Contains(low, "stderr") && !strings.Contains(low, "progress") {
		t.Fatalf("inspect help should mention stderr progress; got:\n%s", resp.Stdout)
	}
	if !strings.Contains(low, "save") {
		t.Fatalf("inspect help should mention SAVE / space saved; got:\n%s", resp.Stdout)
	}
}
```
