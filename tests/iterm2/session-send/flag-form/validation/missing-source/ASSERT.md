## Expected

- Exit 1; stderr mentions expected `--session-id` or `--tab` / `--tab-index`; no SendText.

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
	if !strings.Contains(resp.Stderr, "expected --session-id, or --tab / --tab-index") {
		t.Fatalf("stderr=%q", resp.Stderr)
	}
	if len(resp.SendCalls) != 0 {
		t.Fatalf("SendCalls=%#v", resp.SendCalls)
	}
}
```
