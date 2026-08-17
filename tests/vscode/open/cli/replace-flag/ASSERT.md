## Expected
- CLI exits successfully with `--replace` flag.
- No stderr usage or validation errors.

## Exit Code
- 0

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("unexpected error running kool: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for --replace open, got %d\nstderr: %s", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("unexpected stderr: %s", resp.Stderr)
	}
}
```