## Expected
- Exec hook is called.
- On darwin, command is `open` with vscode:// URI as final argument.
- URI path query decodes to normalized repo path.

## Exit Code
- 0 (no validation error)

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
		t.Fatal("OS opener exec must be called")
	}
	if resp.ExecCommand != "open" {
		t.Fatalf("exec command=%q, want open", resp.ExecCommand)
	}
	if len(resp.ExecArgs) == 0 {
		t.Fatal("exec args must include URI")
	}
	uri := resp.ExecArgs[len(resp.ExecArgs)-1]
	if !strings.HasPrefix(uri, "vscode://xhd2015.open-in-new-window/git-open?path=") {
		t.Fatalf("exec URI=%q", uri)
	}
	parsed, _ := url.Parse(uri)
	if parsed.Query().Get("path") != req.RepoPath {
		t.Fatalf("URI path=%q, want repo %q", parsed.Query().Get("path"), req.RepoPath)
	}
}
```