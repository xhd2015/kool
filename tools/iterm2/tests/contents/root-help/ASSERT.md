## Expected

- Exit 0
- Help lists `contents`

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "contents") {
		t.Fatalf("root help missing contents:\n%s", resp.Stdout)
	}
}
```
