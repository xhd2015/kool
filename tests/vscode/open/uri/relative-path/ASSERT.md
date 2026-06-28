## Expected
- Validation resolves relative path against cwd.
- URI encodes the absolute normalized path.

```go
import (
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ValidateErr != "" {
		t.Fatalf("unexpected validation error: %s", resp.ValidateErr)
	}
	wantAbs, _ := filepath.Abs(filepath.Join(req.WorkingDir, "subdir"))
	if resp.NormalizedPath != wantAbs {
		t.Fatalf("normalized=%q, want %q", resp.NormalizedPath, wantAbs)
	}
	if !strings.HasPrefix(resp.VSCodeURI, "vscode://xhd2015.open-in-new-window/open?path=") {
		t.Fatalf("unexpected URI: %s", resp.VSCodeURI)
	}
	parsed, _ := url.Parse(resp.VSCodeURI)
	if parsed.Query().Get("path") != wantAbs {
		t.Fatalf("URI path=%q, want %q", parsed.Query().Get("path"), wantAbs)
	}
}
```