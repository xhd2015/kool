## Expected
- Validation succeeds.
- URI is `vscode://xhd2015.open-in-new-window/git-open?path=<encoded-abs-path>`.

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
	if !strings.HasPrefix(resp.VSCodeURI, "vscode://xhd2015.open-in-new-window/git-open?path=") {
		t.Fatalf("unexpected URI: %s", resp.VSCodeURI)
	}
	parsed, parseErr := url.Parse(resp.VSCodeURI)
	if parseErr != nil {
		t.Fatalf("invalid URI: %v", parseErr)
	}
	decoded := parsed.Query().Get("path")
	if decoded != resp.NormalizedPath {
		t.Fatalf("decoded path=%q, want normalized %q", decoded, resp.NormalizedPath)
	}
}
```