## Expected Output

```
Note: extension not reachable via IPC; opening via vscode:// URI.
```

## Expected
- Stderr contains IPC unreachable hint.
- OS opener invoked with `vscode://xhd2015.open-in-new-window/git-open?path=...`.
- URI path query decodes to normalized repo path.

## Exit Code
- 0

```go
import (
	"net/url"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ValidateErr != "" {
		t.Fatalf("unexpected open error: %s", resp.ValidateErr)
	}
	if !resp.StderrHint {
		assert.Output(t, resp.Stderr, `
Note: extension not reachable via IPC; opening via vscode:// URI.`)
	}
	if !resp.ExecCalled {
		t.Fatal("OS opener must be called on IPC failure")
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