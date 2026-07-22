## Expected

- Non-zero exit.
- Message mentions unresolved / unknown variable / `${` / unknownToken.

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
		t.Fatalf("unresolved var should error; stdout=%s", resp.Stdout)
	}
	out := strings.ToLower(combinedOut(resp))
	if !strings.Contains(out, "unknown") &&
		!strings.Contains(out, "unresolved") &&
		!strings.Contains(out, "variable") &&
		!strings.Contains(out, "unknowntoken") &&
		!strings.Contains(combinedOut(resp), "${") {
		t.Fatalf("expected unresolved-var error; out:\n%s", combinedOut(resp))
	}
}
```
