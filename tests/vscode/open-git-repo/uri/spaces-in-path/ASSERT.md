## Expected
- URI contains `%20` for spaces (or equivalent proper URL encoding).
- Decoded path matches normalized path with spaces.

```go
import (
	"net/url"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ValidateErr != "" {
		t.Fatalf("unexpected validation error: %s", resp.ValidateErr)
	}
	if !strings.Contains(resp.VSCodeURI, "%20") && !strings.Contains(resp.VSCodeURI, "+") {
		t.Fatalf("URI must encode spaces, got: %s", resp.VSCodeURI)
	}
	parsed, _ := url.Parse(resp.VSCodeURI)
	decoded := parsed.Query().Get("path")
	if decoded != resp.NormalizedPath {
		t.Fatalf("decoded=%q, want %q", decoded, resp.NormalizedPath)
	}
	if !strings.Contains(decoded, " ") {
		t.Fatalf("decoded path should contain space: %q", decoded)
	}
}
```