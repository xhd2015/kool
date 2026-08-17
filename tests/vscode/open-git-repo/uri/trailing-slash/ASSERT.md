## Expected
- Normalized path has no trailing slash.
- URI query uses slash-free path.

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
	if strings.HasSuffix(resp.NormalizedPath, "/") {
		t.Fatalf("normalized path must not have trailing slash: %q", resp.NormalizedPath)
	}
	parsed, _ := url.Parse(resp.VSCodeURI)
	if strings.HasSuffix(parsed.Query().Get("path"), "/") {
		t.Fatalf("URI path must not have trailing slash: %q", parsed.Query().Get("path"))
	}
}
```