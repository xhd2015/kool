## Expected

- Exit 0.
- TOP legacy table present.
- KEEP column is v1.2.0 and PATH contains example.com/foo.
- notice mentions --root.

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
	out := resp.Stdout
	if !strings.Contains(out, "TOP legacy") {
		t.Fatalf("expected TOP legacy table; got:\n%s", out)
	}
	if !strings.Contains(out, "v1.2.0") {
		t.Fatalf("KEEP should be newest v1.2.0; got:\n%s", out)
	}
	if !strings.Contains(out, "example.com/foo") {
		t.Fatalf("expected module path; got:\n%s", out)
	}
	if !strings.Contains(out, "SAVE:") {
		t.Fatalf("expected SAVE line; got:\n%s", out)
	}
	if !strings.Contains(out, "if prune keeps newest") {
		t.Fatalf("SAVE should describe keep-newest prune; got:\n%s", out)
	}
	if !strings.Contains(out, "% of total") && !strings.Contains(out, "<1% of total") {
		t.Fatalf("non-zero SAVE should include percent of total; got:\n%s", out)
	}
	if !strings.Contains(out, "--root") {
		t.Fatalf("expected notice about --root; got:\n%s", out)
	}
}
```
