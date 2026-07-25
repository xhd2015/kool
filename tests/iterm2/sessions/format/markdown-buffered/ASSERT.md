## Expected

- Exit 0.
- Stdout contains `# iTerm2 snapshot`.
- Contains fixture id or host/windows summary.
- Not progressive CLI (`SawW1BeforeLastListTabs` false).

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
	if !strings.Contains(out, "# iTerm2 snapshot") {
		t.Fatalf("missing markdown heading:\n%s", out)
	}
	if !strings.Contains(out, "AAAAAAAA-0000-0000-0000-000000000001") && !strings.Contains(out, "windows") {
		t.Fatalf("markdown missing fixture/session content:\n%s", out)
	}
	if resp.SawW1BeforeLastListTabs {
		t.Fatal("markdown must be buffered (no progressive CLI W1 during ListTabs)")
	}
}
```
