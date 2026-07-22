## Expected

- Non-zero exit.
- Message indicates conflict between new-window and no-new-window (or mutual exclusion).

## Errors

- Window flags mutually exclusive (always, independent of backend).

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
		t.Fatalf("-n and --no-new-window must conflict; stdout=%s", resp.Stdout)
	}
	out := strings.ToLower(combinedOut(resp))
	// Accept several phrasings of mutual exclusion.
	ok := strings.Contains(out, "no-new-window") ||
		strings.Contains(out, "new-window") ||
		strings.Contains(out, "mutually") ||
		strings.Contains(out, "conflict") ||
		strings.Contains(out, "exclusive") ||
		(strings.Contains(out, "both") && strings.Contains(out, "window"))
	if !ok {
		t.Fatalf("expected window flag conflict message; out:\n%s", combinedOut(resp))
	}
}
```
