## Expected

- Non-zero exit.
- Message mentions duplicate and/or id.

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
		t.Fatalf("expected duplicate id error; out:\n%s", out)
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		t.Fatalf("--tab not accepted; out:\n%s", out)
	}
	if !strings.Contains(lower, "duplicate") && !strings.Contains(lower, "duplicated") {
		// also accept generic invalid + id
		if !(strings.Contains(lower, "id") && (strings.Contains(lower, "invalid") || strings.Contains(lower, "conflict") || strings.Contains(lower, "already"))) {
			t.Fatalf("expected duplicate-id error; out:\n%s", out)
		}
	}
}
```
