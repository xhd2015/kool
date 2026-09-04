## Expected

- Non-zero exit.
- Stderr contains `Error:` and unknown/unrecognized.

## Exit Code

- non-zero

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "Error:") {
		t.Fatalf("stderr must contain Error:; got %q", resp.Stderr)
	}
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "unknown") && !strings.Contains(low, "unrecognized") {
		t.Fatalf("stderr should mention unknown/unrecognized; got %q", resp.Stderr)
	}
}
```
