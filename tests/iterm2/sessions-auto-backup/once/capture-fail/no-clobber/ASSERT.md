## Expected

- Exit 0
- warning: on stderr
- File still contains old-sess (not clobbered)

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
		t.Fatalf("capture fail must exit 0; exit=%d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "warning:") {
		t.Fatalf("expected warning: on stderr; stderr=%q", resp.Stderr)
	}
	if resp.Doc == nil {
		t.Fatal("previous backup should still exist")
	}
	if len(resp.Doc.Windows) == 0 || len(resp.Doc.Windows[0].Tabs) == 0 ||
		resp.Doc.Windows[0].Tabs[0].SessionID != "old-sess" {
		t.Fatalf("must not clobber on capture fail; doc=%+v file=%s", resp.Doc, resp.FileJSON)
	}
}
```
