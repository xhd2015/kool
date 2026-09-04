## Expected

- Exit 0.
- Stdout starts with `would remove`.
- example.com/foo@v1.0.0 extracted dir and zip still exist.

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"golang.org/x/mod/module"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "would remove") {
		t.Fatalf("dry-run should say would remove; got:\n%s", resp.Stdout)
	}
	escaped, escErr := module.EscapePath("example.com/foo")
	if escErr != nil {
		t.Fatal(escErr)
	}
	oldDir := filepath.Join(req.ModCache, filepath.FromSlash(escaped)+"@v1.0.0")
	if _, err := os.Stat(oldDir); err != nil {
		t.Fatalf("dry-run must keep %s: %v", oldDir, err)
	}
	oldZip := filepath.Join(req.ModCache, "cache", "download", filepath.FromSlash(escaped), "@v", "v1.0.0.zip")
	if _, err := os.Stat(oldZip); err != nil {
		t.Fatalf("dry-run must keep %s: %v", oldZip, err)
	}
}
```
