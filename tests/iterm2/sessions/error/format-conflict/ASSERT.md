## Expected

- Exit code non-zero.
- Stderr contains `mutually exclusive` (existing ResolveFormat message).

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
		t.Fatalf("want non-zero on format conflict; stderr=%q", resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "mutually exclusive") {
		t.Fatalf("stderr=%q want mutually exclusive", resp.Stderr)
	}
}
```
