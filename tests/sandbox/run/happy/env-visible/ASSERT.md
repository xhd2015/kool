## Expected Output

```
bar
```

## Expected

- Build succeeds; sealed run exit 0.
- Stdout is exactly `bar` (no trailing newline required from printf).

## Exit Code

- sealed run: 0

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
		t.Fatalf("build exit=%d want 0; stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if !resp.RunExecuted {
		t.Fatal("expected sealed binary run")
	}
	if resp.RunExitCode != 0 {
		t.Fatalf("sealed exit=%d want 0; stdout=%q stderr=%q", resp.RunExitCode, resp.RunStdout, resp.RunStderr)
	}
	got := strings.TrimRight(resp.RunStdout, "\n")
	if got != "bar" {
		t.Fatalf("FOO want %q got %q (raw stdout=%q stderr=%q)", "bar", got, resp.RunStdout, resp.RunStderr)
	}
}
```
