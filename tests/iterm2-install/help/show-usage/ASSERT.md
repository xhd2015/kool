## Expected

- Exit 0.
- Stdout documents flag name exactly `--via-open`.
- Stdout mentions user-driven install language: at least one of
  `open`, `Gatekeeper`, `user-driven`, `user open`, `quarantine` (case-insensitive).
- Stdout ends with trailing `\n`.
- No download side effects (help).

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stdout == "" {
		t.Fatal("expected help on stdout")
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("help stdout must end with newline; got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `<contains>
--via-open
</contains>
`)
	low := strings.ToLower(resp.Stdout)
	mentionsUserOpen := strings.Contains(low, "open") ||
		strings.Contains(low, "gatekeeper") ||
		strings.Contains(low, "user-driven") ||
		strings.Contains(low, "user open") ||
		strings.Contains(low, "quarantine")
	if !mentionsUserOpen {
		t.Fatalf("help should mention open/Gatekeeper/user-driven/quarantine; got:\n%s", resp.Stdout)
	}
}
```
