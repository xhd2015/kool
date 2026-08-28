## Expected

- Exit 1; missing text.

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
	if !strings.Contains(resp.Stderr, "missing text") {
		t.Fatalf("stderr=%q", resp.Stderr)
	}
}
```
