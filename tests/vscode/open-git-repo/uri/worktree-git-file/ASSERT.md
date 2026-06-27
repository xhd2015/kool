## Expected
- Validation accepts worktree path.
- URI encodes worktree absolute path.

```go
import (
	"net/url"
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
	if resp.NormalizedPath != req.RepoPath {
		t.Fatalf("normalized=%q, want worktree %q", resp.NormalizedPath, req.RepoPath)
	}
	if !strings.HasPrefix(resp.VSCodeURI, "vscode://xhd2015.open-in-new-window/git-open?path=") {
		t.Fatalf("unexpected URI: %s", resp.VSCodeURI)
	}
	parsed, _ := url.Parse(resp.VSCodeURI)
	if parsed.Query().Get("path") != req.RepoPath {
		t.Fatalf("URI path=%q, want %q", parsed.Query().Get("path"), req.RepoPath)
	}
}
```