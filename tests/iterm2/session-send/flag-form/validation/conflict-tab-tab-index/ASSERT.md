## Expected

- Exit 1; `--tab and --tab-index cannot be specified together`.

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
	if !strings.Contains(resp.Stderr, "--tab and --tab-index cannot be specified together") {
		t.Fatalf("stderr=%q", resp.Stderr)
	}
}
```
