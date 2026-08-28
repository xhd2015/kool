## Expected

- Exit 1; `--session-id cannot be combined with --tab/--tab-index`.

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
	if resp.ExitCode != 1 {
		t.Fatalf("exit=%d", resp.ExitCode)
	}
	if !strings.Contains(resp.Stderr, "--session-id cannot be combined") {
		t.Fatalf("stderr=%q", resp.Stderr)
	}
	if len(resp.SendCalls) != 0 {
		t.Fatal("SendText must not run")
	}
}
```
