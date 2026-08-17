## Expected

- Exit 1; no captured script.

## Exit Code

- 1

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected failure, stderr=%s", resp.Stderr)
	}
	if resp.CapturedScript != "" {
		t.Fatal("osascript should not run for missing path")
	}
}
```