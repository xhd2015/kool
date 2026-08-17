## Expected

- Non-zero exit.
- Message mentions not found / unknown / missing.

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
		t.Fatalf("show missing should error; stdout=%s", resp.Stdout)
	}
	out := strings.ToLower(combinedOut(resp))
	if !strings.Contains(out, "not found") &&
		!strings.Contains(out, "unknown") &&
		!strings.Contains(out, "no such") &&
		!strings.Contains(out, "missing") {
		t.Fatalf("expected not-found error; out:\n%s", combinedOut(resp))
	}
}
```
