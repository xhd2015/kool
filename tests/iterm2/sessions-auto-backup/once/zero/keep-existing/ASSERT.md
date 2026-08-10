## Expected

- Exit 0
- 0 critical message (previous backup kept intent)
- File still contains old-sess

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
	out := resp.Stdout + "\n" + resp.Stderr
	if !strings.Contains(out, "0 critical") {
		t.Fatalf("expected 0 critical message:\n%s", out)
	}
	if resp.Doc == nil {
		t.Fatal("previous backup file should still exist")
	}
	if resp.Doc.Windows == nil || len(resp.Doc.Windows) == 0 ||
		len(resp.Doc.Windows[0].Tabs) == 0 ||
		resp.Doc.Windows[0].Tabs[0].SessionID != "old-sess" {
		t.Fatalf("must not clobber existing backup; doc=%+v file=%s", resp.Doc, resp.FileJSON)
	}
}
```
