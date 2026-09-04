## Expected

- Exit 0.
- Stdout mentions `prune` and `--dry-run`.
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
	if !strings.Contains(low, "prune") {
		t.Fatalf("prune help should mention prune; got:\n%s", resp.Stdout)
	}
	if !strings.Contains(low, "--dry-run") {
		t.Fatalf("prune help should mention --dry-run; got:\n%s", resp.Stdout)
	}
}
```
