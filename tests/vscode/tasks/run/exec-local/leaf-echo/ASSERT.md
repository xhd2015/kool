## Expected

- Exit 0.
- Stdout (or combined output) contains marker `KOOL_TASKS_P2_HELLO`.
- Real execution (not dry-run-only plan text without the echo output).

## Side Effects

- Shell child process ran under local backend (no iTerm required).

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
		t.Fatalf("local leaf-echo exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := combinedOut(resp)
	if !strings.Contains(out, "KOOL_TASKS_P2_HELLO") {
		t.Fatalf("expected echo marker KOOL_TASKS_P2_HELLO; out:\n%s", out)
	}
}
```
