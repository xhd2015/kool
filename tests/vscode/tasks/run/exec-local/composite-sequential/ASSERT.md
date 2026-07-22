## Expected

- Exit 0.
- Combined output contains both `KOOL_TASKS_P2_STEP_ONE` and `KOOL_TASKS_P2_STEP_TWO`.
- Composite deps expanded then executed (local may sequentialize parallel plan siblings).

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
		t.Fatalf("local composite exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := combinedOut(resp)
	for _, want := range []string{"KOOL_TASKS_P2_STEP_ONE", "KOOL_TASKS_P2_STEP_TWO"} {
		if !strings.Contains(out, want) {
			t.Fatalf("composite sequential missing %q; out:\n%s", want, out)
		}
	}
}
```
