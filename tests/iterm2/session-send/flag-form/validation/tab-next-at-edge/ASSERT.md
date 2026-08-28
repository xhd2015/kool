## Expected

- Exit 1; no tab to the right; no SendText.

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
		t.Fatalf("exit=%d stderr=%s", resp.ExitCode, resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "no tab to the right") {
		t.Fatalf("stderr=%q", resp.Stderr)
	}
	if len(resp.SendCalls) != 0 {
		t.Fatal("SendText must not run")
	}
}
```
