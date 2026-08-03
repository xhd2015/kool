## Expected

- Exit 0
- Stdout contains Saved
- FileJSON includes a `"space"` key under windows (value may be 0)
- File is valid checkpoint (version 1, not consumed)

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "Saved") {
		t.Fatalf("stdout:\n%s", resp.Stdout)
	}
	if resp.FileJSON == "" {
		t.Fatal("missing FileJSON")
	}
	// Always emit space when not --ignore-macos-space (including 0).
	if !strings.Contains(resp.FileJSON, `"space"`) {
		t.Fatalf("checkpoint must emit \"space\" when not ignore:\n%s", resp.FileJSON)
	}
	if resp.Doc == nil {
		t.Fatal("missing parsed Doc")
	}
	if resp.Doc.Version != 1 {
		t.Fatalf("version=%d", resp.Doc.Version)
	}
	if resp.Doc.IsConsumed() {
		t.Fatal("restored_at should be empty after save")
	}
}
```
