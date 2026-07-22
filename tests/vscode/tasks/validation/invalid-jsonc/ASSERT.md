## Expected

- Non-zero exit.
- Message mentions parse / invalid / JSON / syntax.

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
		t.Fatalf("invalid JSONC should error; stdout=%s", resp.Stdout)
	}
	out := strings.ToLower(combinedOut(resp))
	if !strings.Contains(out, "parse") &&
		!strings.Contains(out, "invalid") &&
		!strings.Contains(out, "json") &&
		!strings.Contains(out, "syntax") &&
		!strings.Contains(out, "unexpected") {
		t.Fatalf("expected parse/invalid error; out:\n%s", combinedOut(resp))
	}
}
```
