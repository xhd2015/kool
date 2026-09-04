## Expected

- Exit 1.
- Stderr contains `Error:` and mentions the path or "directory".

## Exit Code

- 1

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 1 {
		t.Fatalf("exit=%d want 1; stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "Error:") {
		t.Fatalf("stderr must contain Error:; got %q", resp.Stderr)
	}
	base := filepath.Base(req.ModCache)
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "directory") && !strings.Contains(resp.Stderr, base) {
		t.Fatalf("stderr should mention directory or path %q; got %q", base, resp.Stderr)
	}
}
```
