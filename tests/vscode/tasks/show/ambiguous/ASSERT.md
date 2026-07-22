## Expected

- Non-zero exit.
- Message mentions ambiguous / multiple; prefer listing Alpha One and Alpha Two.

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
		t.Fatalf("show ambiguous should error; stdout=%s", resp.Stdout)
	}
	out := strings.ToLower(combinedOut(resp))
	if !strings.Contains(out, "ambiguous") &&
		!strings.Contains(out, "multiple") &&
		!strings.Contains(out, "more than one") {
		// still require match names somewhere
		raw := combinedOut(resp)
		if !strings.Contains(raw, "Alpha One") || !strings.Contains(raw, "Alpha Two") {
			t.Fatalf("expected ambiguous error mentioning matches; out:\n%s", raw)
		}
	}
}
```
