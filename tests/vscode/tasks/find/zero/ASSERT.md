## Expected

- Non-zero exit.
- Message indicates no match / not found.

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
		t.Fatalf("find zero should error; stdout=%s", resp.Stdout)
	}
	out := strings.ToLower(combinedOut(resp))
	if !strings.Contains(out, "not found") &&
		!strings.Contains(out, "no match") &&
		!strings.Contains(out, "no task") &&
		!strings.Contains(out, "zzzz") {
		t.Fatalf("expected zero-match error; out:\n%s", combinedOut(resp))
	}
}
```
