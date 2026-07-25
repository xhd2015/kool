## Expected

- Exit 0.
- Stdout contains `<html` (case-insensitive acceptable via lower check).
- Fixture session id present.
- Buffered: `SawW1BeforeLastListTabs` false.

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
	if !strings.Contains(strings.ToLower(out), "<html") {
		t.Fatalf("missing <html:\n%s", out)
	}
	if !strings.Contains(out, "AAAAAAAA-0000-0000-0000-000000000001") {
		t.Fatalf("html missing fixture session id:\n%s", out)
	}
	if resp.SawW1BeforeLastListTabs {
		t.Fatal("html must be buffered (no progressive CLI W1 during ListTabs)")
	}
}
```
