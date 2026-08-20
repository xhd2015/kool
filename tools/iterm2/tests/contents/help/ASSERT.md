## Expected

- Exit 0
- Stdout mentions Usage and contents
- Stderr empty

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
	if resp.Stderr != "" {
		t.Fatalf("stderr=%q", resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "Usage:") && !strings.Contains(resp.Stdout, "contents") {
		t.Fatalf("stdout=%q", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "contents") {
		t.Fatalf("missing contents in help: %q", resp.Stdout)
	}
}
```
