## Expected

- Non-zero exit.
- Message mentions force and/or save.

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
	out := combinedOut(resp)
	if resp.ExitCode == 0 {
		t.Fatalf("--force without --save must error; out:\n%s", out)
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		t.Fatalf("--force/--tab not accepted; out:\n%s", out)
	}
	if !strings.Contains(lower, "force") && !strings.Contains(lower, "save") {
		t.Fatalf("expected force/save error message; out:\n%s", out)
	}
}
```
