## Expected

- Non-zero exit (prefer 1).
- Message mentions tasks.json and/or not found / missing.

## Errors

- Missing workspace tasks file after parent walk.

## Exit Code

- ≠ 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero when tasks.json missing; stdout=%s", resp.Stdout)
	}
	out := strings.ToLower(combinedOut(resp))
	if !strings.Contains(out, "tasks.json") &&
		!strings.Contains(out, "not found") &&
		!strings.Contains(out, "missing") &&
		!strings.Contains(out, "no task") {
		t.Fatalf("expected missing tasks.json error; out:\n%s", combinedOut(resp))
	}
}
```
