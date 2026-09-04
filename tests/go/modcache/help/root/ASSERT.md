## Expected

- Exit 0.
- Stdout mentions `inspect` and `prune`.
- Stdout ends with a trailing newline.

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
	if !strings.Contains(low, "inspect") || !strings.Contains(low, "prune") {
		t.Fatalf("root help should mention inspect and prune; got:\n%s", resp.Stdout)
	}
}
```
