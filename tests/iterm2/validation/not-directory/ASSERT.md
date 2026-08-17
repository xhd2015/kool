## Expected

- Exit 1; stderr mentions not a directory or similar.

## Exit Code

- 1

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
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(resp.Stderr, "directory") {
		t.Fatalf("stderr=%q", resp.Stderr)
	}
}
```