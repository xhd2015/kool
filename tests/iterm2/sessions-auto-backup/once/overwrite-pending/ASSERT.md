## Expected

- Exit 0 (always overwrite auto file — unlike manual save pending-non-tty)
- Stdout Saved
- File no longer contains old-sess; has grok fixture session

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
		t.Fatalf("auto-backup must overwrite pending without TTY error; exit=%d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	low := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	if strings.Contains(low, "cannot") && strings.Contains(low, "tty") {
		t.Fatalf("must not require TTY for auto overwrite; stderr=%q", resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "Saved") {
		t.Fatalf("stdout missing Saved:\n%s", resp.Stdout)
	}
	if resp.Doc == nil {
		t.Fatal("missing overwritten doc")
	}
	if strings.Contains(resp.FileJSON, "old-sess") {
		t.Fatalf("pending seed must be overwritten; still has old-sess:\n%s", resp.FileJSON)
	}
	if !strings.Contains(resp.FileJSON, fixtureGrokSessionID) {
		t.Fatalf("overwritten file should contain fixture grok session:\n%s", resp.FileJSON)
	}
	if resp.Doc.Summary.Sessions < 1 {
		t.Fatalf("sessions=%d", resp.Doc.Summary.Sessions)
	}
}
```
