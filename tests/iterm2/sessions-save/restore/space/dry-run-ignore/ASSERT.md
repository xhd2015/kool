## Expected

- Exit 0
- Would restore present
- No `space N (Desktop …)` placement lines
- Not stamped

## Exit Code

- 0

```go
import (
	"regexp"
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
	if !strings.Contains(out, "Would restore") {
		t.Fatalf("missing Would restore:\n%s", out)
	}
	re := regexp.MustCompile(`space\s+\d+\s+\(Desktop\s+\d+\)`)
	if re.MatchString(out) {
		t.Fatalf("--ignore-macos-space dry-run must omit space placement lines:\n%s", out)
	}
	if resp.Doc == nil || resp.Doc.IsConsumed() {
		t.Fatal("dry-run must not stamp restored_at")
	}
}
```
