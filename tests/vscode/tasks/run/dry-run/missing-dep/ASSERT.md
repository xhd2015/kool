## Expected

- Non-zero exit.
- Message mentions missing / not found dependency / No Such Task.

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
		t.Fatalf("missing dep should error; stdout=%s", resp.Stdout)
	}
	out := strings.ToLower(combinedOut(resp))
	raw := combinedOut(resp)
	if !strings.Contains(out, "depend") &&
		!strings.Contains(out, "not found") &&
		!strings.Contains(out, "missing") &&
		!strings.Contains(raw, "No Such Task") {
		t.Fatalf("expected missing dependency error; out:\n%s", raw)
	}
}
```
