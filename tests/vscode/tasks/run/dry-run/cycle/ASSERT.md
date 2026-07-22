## Expected

- Non-zero exit.
- Message mentions cycle / circular / loop.

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
		t.Fatalf("cycle should error; stdout=%s", resp.Stdout)
	}
	out := strings.ToLower(combinedOut(resp))
	if !strings.Contains(out, "cycle") &&
		!strings.Contains(out, "circular") &&
		!strings.Contains(out, "loop") {
		t.Fatalf("expected cycle error; out:\n%s", combinedOut(resp))
	}
}
```
