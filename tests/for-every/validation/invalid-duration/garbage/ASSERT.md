## Expected

- Non-zero exit before any successful loop iteration.
- Stderr mentions duration (invalid / parse error).

## Errors

- Invalid duration string.

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
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for garbage duration; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	low := strings.ToLower(resp.Stderr)
	if strings.Contains(low, "unrecognized command") {
		t.Fatalf("for-every must be routed to its handler; got %q", resp.Stderr)
	}
	if !strings.Contains(low, "duration") {
		t.Fatalf("stderr should mention duration; got %q", resp.Stderr)
	}
}
```
