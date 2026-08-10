## Expected

- Exit 0
- Stdout Saved
- File version=1, restored_at empty, grok+mark sessions
- Source is kool-iterm2-sessions-auto (or contains sessions-auto)

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
		t.Fatalf("stdout missing Saved:\n%s", resp.Stdout)
	}
	if resp.Doc == nil {
		t.Fatalf("missing file/doc; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
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
	// Prefer source kool-iterm2-sessions-auto (locked product preference).
	src := resp.Doc.Source
	if src != "kool-iterm2-sessions-auto" && !strings.Contains(src, "sessions-auto") {
		t.Fatalf("source=%q want kool-iterm2-sessions-auto", src)
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
	joined := strings.Join(kinds, ",")
	if !strings.Contains(joined, "grok") || !strings.Contains(joined, "mark") {
		t.Fatalf("kinds=%v", kinds)
	}
}
```
