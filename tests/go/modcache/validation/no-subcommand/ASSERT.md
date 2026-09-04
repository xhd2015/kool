## Expected

- Non-zero exit.
- Stderr mentions inspect and/or prune and/or help.

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
		t.Fatalf("expected non-zero for bare modcache; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "inspect") && !strings.Contains(low, "prune") && !strings.Contains(low, "help") {
		t.Fatalf("stderr should hint inspect/prune/help; got %q", resp.Stderr)
	}
}
```
