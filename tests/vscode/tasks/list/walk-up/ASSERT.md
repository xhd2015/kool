## Expected

- Exit 0.
- Labels from parent tasks.json appear (Compile / Serve / Build All).
- Prefer workspace path points at parent root (not nested).

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("walk-up list exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	if !strings.Contains(out, "Compile") {
		t.Fatalf("walk-up should list parent Compile; out:\n%s", out)
	}
	// Prefer workspace root mentioned (basename of WorkingDir)
	base := filepath.Base(req.WorkingDir)
	if base != "" && !strings.Contains(out, base) && !strings.Contains(out, req.WorkingDir) {
		t.Logf("list ok but workspace path not shown (prefer): %s", out)
	}
}
```
