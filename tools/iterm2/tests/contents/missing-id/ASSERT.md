## Expected

- Exit 1
- Stderr has `Error:` and missing session-id
- Help on stderr

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 1 {
		t.Fatalf("exit %d", resp.ExitCode)
	}
	if !strings.Contains(resp.Stderr, "Error:") {
		t.Fatalf("stderr=%q", resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "missing session-id") {
		t.Fatalf("stderr=%q", resp.Stderr)
	}
}
```
