## Expected

- Non-zero exit.
- Message mentions unresolved / unknown variable / `${` / unknownToken.

## Exit Code

- ≠ 0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
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
