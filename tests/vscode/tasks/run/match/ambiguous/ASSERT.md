## Expected

- Non-zero exit.
- Ambiguous / multiple; prefer listing both Alpha tasks.

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
		t.Fatalf("ambiguous run match should error; stdout=%s", resp.Stdout)
	}
	out := strings.ToLower(combinedOut(resp))
	raw := combinedOut(resp)
	if !strings.Contains(out, "ambiguous") &&
		!strings.Contains(out, "multiple") &&
		!strings.Contains(out, "more than one") {
		if !strings.Contains(raw, "Alpha One") || !strings.Contains(raw, "Alpha Two") {
			t.Fatalf("expected ambiguous match error; out:\n%s", raw)
		}
	}
}
```
