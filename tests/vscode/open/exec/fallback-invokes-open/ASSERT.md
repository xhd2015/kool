## Expected
- Exec hook is called after IPC failure.
- On darwin, command is `open` with vscode:// URI as final argument.
- URI path query decodes to normalized directory path.
- URI omits `replace=` query param (default open).

## Exit Code
- 0

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
		t.Fatalf("unexpected open error: %s", resp.ValidateErr)
	}
	if !resp.ExecCalled {
		t.Fatal("OS opener exec must be called on IPC failure")
	}
	if resp.ExecCommand != "open" {
		t.Fatalf("exec command=%q, want open", resp.ExecCommand)
	}
	if len(resp.ExecArgs) == 0 {
		t.Fatal("exec args must include URI")
	}
	uri := resp.ExecArgs[len(resp.ExecArgs)-1]
	if !strings.HasPrefix(uri, "vscode://xhd2015.open-in-new-window/open?path=") {
		t.Fatalf("exec URI=%q", uri)
	}
	parsed, _ := url.Parse(uri)
	if parsed.Query().Get("path") != req.DirPath {
		t.Fatalf("URI path=%q, want dir %q", parsed.Query().Get("path"), req.DirPath)
	}
	if _, ok := parsed.Query()["replace"]; ok {
		t.Fatalf("URI must not include replace query for default open, got %q", parsed.Query().Get("replace"))
	}
}
```