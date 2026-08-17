## Expected

- Exit 0 after exactly one successful iteration.
- Bare integer duration accepted (no invalid-duration error).

## Exit Code

- 0

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
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	low := strings.ToLower(resp.Stderr)
	if strings.Contains(low, "invalid duration") {
		t.Fatalf("bare int duration should parse; stderr=%q", resp.Stderr)
	}
}
```
