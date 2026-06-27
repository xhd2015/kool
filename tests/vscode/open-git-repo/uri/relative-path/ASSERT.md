## Expected
- Normalized path is absolute (under working dir).
- URI encodes the absolute resolved path.

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
	expectedAbs := filepath.Join(req.WorkingDir, "repo")
	if resp.NormalizedPath != expectedAbs {
		t.Fatalf("normalized=%q, want %q", resp.NormalizedPath, expectedAbs)
	}
	if !strings.HasPrefix(resp.VSCodeURI, "vscode://xhd2015.open-in-new-window/git-open?path=") {
		t.Fatalf("unexpected URI: %s", resp.VSCodeURI)
	}
	parsed, _ := url.Parse(resp.VSCodeURI)
	if parsed.Query().Get("path") != expectedAbs {
		t.Fatalf("URI path query=%q, want %q", parsed.Query().Get("path"), expectedAbs)
	}
}
```