## Expected

- Exit 1
- Stderr `Error: session not found: <id>`

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
	if !strings.HasPrefix(strings.TrimSpace(resp.Stderr), "Error:") {
		t.Fatalf("stderr=%q", resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "session not found") {
		t.Fatalf("stderr=%q", resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "00000000-0000-0000-0000-000000000000") {
		t.Fatalf("stderr=%q", resp.Stderr)
	}
}
```
