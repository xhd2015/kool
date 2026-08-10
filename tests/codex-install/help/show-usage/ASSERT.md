## Expected

- Exit 0.
- Stdout documents flag names exactly `--dry-run` and `--check-update`.
- Stdout mentions install / codex usage (case-insensitive `codex` or `install`).
- Stdout ends with trailing `\n`.
- No shell mutation (help path).

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
	assertNoError(t, err)
	assertExitZero(t, resp)
	if resp.Stdout == "" {
		t.Fatal("expected help on stdout")
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("help stdout must end with newline; got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `<contains>
--dry-run
--check-update
</contains>
`)
	low := strings.ToLower(resp.Stdout)
	if !strings.Contains(low, "codex") && !strings.Contains(low, "install") {
		t.Fatalf("help should mention codex/install; got:\n%s", resp.Stdout)
	}
	assertNoShell(t, resp)
}
```
