## Expected

- Exit 0
- Stdout contains Saved and 1 window (or 1 critical session)
- FileJSON has `"filter"` with spaces containing 0
- Doc has one window on space 0
- Stderr contains skipped windows not matching --spaces

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
	if !strings.Contains(resp.Stderr, "skipped") || !strings.Contains(resp.Stderr, "--spaces") {
		t.Fatalf("expected skip warning; stderr=%q", resp.Stderr)
	}
	if resp.Doc == nil {
		t.Fatal("missing Doc")
	}
	if resp.Doc.Filter == nil || len(resp.Doc.Filter.Spaces) != 1 || resp.Doc.Filter.Spaces[0] != 0 {
		t.Fatalf("filter=%+v", resp.Doc.Filter)
	}
	if resp.Doc.Summary.Windows != 1 || resp.Doc.Summary.Sessions != 1 {
		t.Fatalf("summary=%+v", resp.Doc.Summary)
	}
	if len(resp.Doc.Windows) != 1 || resp.Doc.Windows[0].Space != 0 {
		t.Fatalf("windows=%+v", resp.Doc.Windows)
	}
	if !strings.Contains(resp.FileJSON, `"filter"`) {
		t.Fatalf("FileJSON missing filter:\n%s", resp.FileJSON)
	}
}
```
