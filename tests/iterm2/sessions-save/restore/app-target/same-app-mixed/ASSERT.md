## Expected

- Exit 0
- Would restore both resume cmds
- Plan shows create **`app`** for system **and** home
- not stamped

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
	out := resp.Stdout
	if !strings.Contains(out, "Would restore") {
		t.Fatalf("missing Would restore:\n%s", out)
	}
	if !strings.Contains(out, "grok --resume "+fixtureGrokSessionID) {
		t.Fatalf("missing grok resume:\n%s", out)
	}
	if !strings.Contains(out, "mark '"+fixtureMarkMessage+"'") &&
		!strings.Contains(out, fixtureMarkMessage) {
		t.Fatalf("missing mark resume/message:\n%s", out)
	}
	if !hasSameAppCreateLineFor(out, fixtureAppSystem) {
		t.Fatalf("mixed --same-app must show system create app line:\n%s", out)
	}
	if !hasSameAppCreateLineFor(out, fixtureAppHome) {
		t.Fatalf("mixed --same-app must show home create app line:\n%s", out)
	}
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```
