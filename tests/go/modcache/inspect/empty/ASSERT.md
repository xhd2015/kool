## Expected

- Exit 0.
- Stdout has GOMODCACHE and LEGACY 0 versions.

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
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "GOMODCACHE:") {
		t.Fatalf("expected GOMODCACHE label; got:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "LEGACY:") || !strings.Contains(resp.Stdout, "0 versions") {
		t.Fatalf("empty cache should report 0 legacy versions; got:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "SAVE:") || !strings.Contains(resp.Stdout, "0B") {
		t.Fatalf("empty cache should report SAVE: 0B; got:\n%s", resp.Stdout)
	}
}
```
