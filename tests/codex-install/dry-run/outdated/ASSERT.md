## Expected

- Exit 0.
- Stdout mentions dry-run and update plan.
- Stdout shows local `0.1.0` and latest `0.2.0`.
- Stdout includes `install.UpdateCmd` (`codex update`) or clear “update” wording
  with the command.
- `ShellCalls` empty.

## Side Effects

- No `RunShell` calls.
- FetchLatest is allowed (≥1) because bin is present.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
	codexinstall "github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/install"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	assertExitZero(t, resp)
	assertNoShell(t, resp)
	assert.Output(t, resp.Stdout, `<contains>
dry-run
0.1.0
0.2.0
</contains>
`)
	low := strings.ToLower(resp.Stdout)
	if !strings.Contains(low, "update") {
		t.Fatalf("dry-run outdated plan should mention update; got:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, codexinstall.UpdateCmd) &&
		!strings.Contains(resp.Stdout, "codex update") {
		t.Fatalf("dry-run outdated plan should include UpdateCmd %q; got:\n%s",
			codexinstall.UpdateCmd, resp.Stdout)
	}
}
```
