## Expected

- Exit 0.
- Stdout mentions `package-table` (and ideally `coverage`).
- Stdout ends with a trailing newline.
- No profile work required.

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
	if resp.Stdout == "" {
		t.Fatal("expected help on stdout")
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("help stdout must end with newline; got %q", resp.Stdout)
	}
	low := strings.ToLower(resp.Stdout)
	if !strings.Contains(low, "package-table") {
		t.Fatalf("root help should mention package-table; got:\n%s", resp.Stdout)
	}
}
```
