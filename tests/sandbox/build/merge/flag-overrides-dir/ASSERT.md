## Expected

- Build exit 0; binary size > 0.
- Inspect ran with exit 0.
- Inspect stdout mentions packed path `shared.txt` and env key `SHARED`.
- Inspect stdout must **not** contain env secret values `from-flag` or `from-dir`
  (keys only policy).
- File path present once (single winning path).

## Exit Code

- 0 (build)

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
		t.Fatalf("build exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !resp.OutputExists || resp.OutputSize <= 0 {
		t.Fatalf("expected sealed binary; exists=%v size=%d", resp.OutputExists, resp.OutputSize)
	}
	if !resp.InspectRan {
		t.Fatal("expected AfterBuildInspect to run")
	}
	if resp.InspectExitCode != 0 {
		t.Fatalf("inspect exit=%d stderr=%q stdout=%q", resp.InspectExitCode, resp.InspectStderr, resp.InspectStdout)
	}
	insp := resp.InspectStdout
	if !strings.Contains(insp, "shared.txt") {
		t.Fatalf("inspect should list path shared.txt; got %q", insp)
	}
	if !strings.Contains(insp, "SHARED") {
		t.Fatalf("inspect should list env key SHARED; got %q", insp)
	}
	// Never leak env secret values on inspect.
	if strings.Contains(insp, "from-flag") || strings.Contains(insp, "from-dir") {
		t.Fatalf("inspect must not print env secret values; got %q", insp)
	}
}
```
