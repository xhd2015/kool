## Expected

- Exit 0
- File version=1, restored_at null/absent, 2 sessions, kinds grok+mark
- Resume cmds present

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
		t.Fatalf("exit=%d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "Saved") {
		t.Fatalf("stdout:\n%s", resp.Stdout)
	}
	if resp.Doc == nil {
		t.Fatal("missing file/doc")
	}
	if resp.Doc.Version != 1 {
		t.Fatalf("version=%d", resp.Doc.Version)
	}
	if resp.Doc.IsConsumed() {
		t.Fatal("restored_at should be empty")
	}
	if resp.Doc.Summary.Sessions != 2 {
		t.Fatalf("sessions=%d", resp.Doc.Summary.Sessions)
	}
	if resp.Doc.Summary.ByKind["grok"] != 1 || resp.Doc.Summary.ByKind["mark"] != 1 {
		t.Fatalf("by_kind=%v", resp.Doc.Summary.ByKind)
	}
	var kinds []string
	for _, w := range resp.Doc.Windows {
		for _, tab := range w.Tabs {
			kinds = append(kinds, tab.Kind)
			if tab.ResumeCmd == "" {
				t.Fatal("empty resume_cmd")
			}
		}
	}
	if !strings.Contains(strings.Join(kinds, ","), "grok") || !strings.Contains(strings.Join(kinds, ","), "mark") {
		t.Fatalf("kinds=%v", kinds)
	}
}
```
