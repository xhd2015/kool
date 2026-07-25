## Expected

- Non-zero exit
- Error mentions TTY / not restored / overwrite
- File still has old-sess

## Exit Code

- ≠ 0

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
	if resp.ExitCode == 0 {
		t.Fatalf("expected failure; stdout=%q", resp.Stdout)
	}
	low := strings.ToLower(resp.Stderr + resp.Stdout)
	if !strings.Contains(low, "tty") && !strings.Contains(low, "not restored") && !strings.Contains(low, "overwrite") {
		t.Fatalf("error should mention TTY/overwrite:\n%s", resp.Stderr)
	}
	if resp.Doc == nil || len(resp.Doc.Windows) == 0 || resp.Doc.Windows[0].Tabs[0].SessionID != "old-sess" {
		t.Fatalf("file should remain old-sess; doc=%+v", resp.Doc)
	}
}
```
