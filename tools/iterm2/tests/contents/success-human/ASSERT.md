## Expected

- Exit 0
- Stdout is exactly the pane plus a trailing newline
- No app header on stdout

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
	if resp.Stdout != "❯ hello\n" {
		t.Fatalf("stdout=%q", resp.Stdout)
	}
	if strings.Contains(resp.Stdout, "app") {
		t.Fatalf("human stdout should not mention app: %q", resp.Stdout)
	}
}
```
