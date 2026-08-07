## Expected

- Build exit 0; sealed binary exists with size > 0.
- Inspect ran with exit 0.
- Inspect stdout mentions a runtime-load-devbox section/label and lists both
  absolute paths from the request (order preserved when product prints a list).
- Stdout of build ends with `\n`.

## Exit Code

- 0 (build)

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("build exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !resp.OutputExists || resp.OutputSize <= 0 {
		t.Fatalf("expected sealed binary; exists=%v size=%d path=%q", resp.OutputExists, resp.OutputSize, resp.OutputPath)
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("build stdout must end with newline; got %q", resp.Stdout)
	}
	if len(req.RuntimeLoadDevbox) < 2 {
		t.Fatalf("expected ≥2 RuntimeLoadDevbox paths in request; got %v", req.RuntimeLoadDevbox)
	}
	if !resp.InspectRan {
		t.Fatal("expected AfterBuildInspect to run")
	}
	if resp.InspectExitCode != 0 {
		t.Fatalf("inspect exit=%d stderr=%q stdout=%q", resp.InspectExitCode, resp.InspectStderr, resp.InspectStdout)
	}
	insp := resp.InspectStdout
	low := strings.ToLower(insp)
	// Section label: hyphen, underscore, or words without separators.
	if !strings.Contains(low, "runtime-load-devbox") &&
		!strings.Contains(low, "runtime_load_devbox") &&
		!strings.Contains(low, "runtime load devbox") {
		t.Fatalf("inspect should label runtime-load-devbox section; got %q", insp)
	}
	for i, p := range req.RuntimeLoadDevbox {
		if !strings.Contains(insp, p) {
			t.Fatalf("inspect should list sealed path[%d]=%q; got %q", i, p, insp)
		}
	}
	// Order: first path should appear before second when both present.
	i0 := strings.Index(insp, req.RuntimeLoadDevbox[0])
	i1 := strings.Index(insp, req.RuntimeLoadDevbox[1])
	if i0 < 0 || i1 < 0 || i0 > i1 {
		t.Fatalf("inspect should list sealed paths in pack order; i0=%d i1=%d out=%q", i0, i1, insp)
	}
}
```
