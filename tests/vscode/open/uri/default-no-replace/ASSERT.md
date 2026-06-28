## Expected
- Validation succeeds.
- URI is `vscode://xhd2015.open-in-new-window/open?path=<encoded>`.
- Query string does not contain `replace` parameter.

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
	want := expectedOpenURI(resp.NormalizedPath, false)
	if resp.VSCodeURI != want {
		t.Fatalf("URI=%q, want %q", resp.VSCodeURI, want)
	}
	if !strings.HasPrefix(resp.VSCodeURI, "vscode://xhd2015.open-in-new-window/open?path=") {
		t.Fatalf("unexpected URI prefix: %s", resp.VSCodeURI)
	}
	parsed, parseErr := url.Parse(resp.VSCodeURI)
	if parseErr != nil {
		t.Fatalf("invalid URI: %v", parseErr)
	}
	if _, ok := parsed.Query()["replace"]; ok {
		t.Fatalf("URI must not include replace query, got %q", parsed.Query().Get("replace"))
	}
}
```