## Expected
- Validation succeeds.
- URI includes `replace=true` query parameter after encoded path.

```go
import (
	"net/url"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ValidateErr != "" {
		t.Fatalf("unexpected validation error: %s", resp.ValidateErr)
	}
	want := expectedOpenURI(resp.NormalizedPath, true)
	if resp.VSCodeURI != want {
		t.Fatalf("URI=%q, want %q", resp.VSCodeURI, want)
	}
	parsed, parseErr := url.Parse(resp.VSCodeURI)
	if parseErr != nil {
		t.Fatalf("invalid URI: %v", parseErr)
	}
	if parsed.Query().Get("replace") != "true" {
		t.Fatalf("replace query=%q, want true", parsed.Query().Get("replace"))
	}
	if parsed.Query().Get("path") != resp.NormalizedPath {
		t.Fatalf("decoded path=%q, want normalized %q", parsed.Query().Get("path"), resp.NormalizedPath)
	}
}
```