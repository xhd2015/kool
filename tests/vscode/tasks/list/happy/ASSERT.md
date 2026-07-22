## Expected

- Exit 0.
- Stdout includes all labels: `Build All`, `Compile`, `Serve`.
- Prefer type hints: composite / shell / process; BG for Serve; dep count for Build All.
- Prefer labels appear in sorted order (Build All before Compile before Serve).
- Footer may mention task count `3` and workspace path.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("list exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	for _, label := range []string{"Build All", "Compile", "Serve"} {
		if !strings.Contains(out, label) {
			t.Fatalf("list missing label %q; out:\n%s", label, out)
		}
	}
	// Sorted order: Build All index < Compile < Serve
	iBuild := strings.Index(out, "Build All")
	iCompile := strings.Index(out, "Compile")
	iServe := strings.Index(out, "Serve")
	if iBuild < 0 || iCompile < 0 || iServe < 0 || !(iBuild < iCompile && iCompile < iServe) {
		t.Fatalf("labels should be sorted Build All, Compile, Serve; out:\n%s", out)
	}
	lower := strings.ToLower(out)
	// Soft type hints
	if !strings.Contains(lower, "shell") && !strings.Contains(lower, "composite") &&
		!strings.Contains(lower, "process") {
		t.Logf("list has labels but no type column hints: %s", out)
	}
}
```
