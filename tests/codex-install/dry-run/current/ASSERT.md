## Expected

- Exit 0.
- Stdout mentions dry-run and a noop / up-to-date / current outcome.
- Stdout may include version `0.147.0`.
- Must **not** claim would install or would update as the planned action
  (no `InstallCmd` / no forced update command as the plan).
- `ShellCalls` empty.

## Side Effects

- No `RunShell` calls.

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
</contains>
`)
	low := strings.ToLower(resp.Stdout)
	noopish := strings.Contains(low, "noop") ||
		strings.Contains(low, "up to date") ||
		strings.Contains(low, "up-to-date") ||
		strings.Contains(low, "current") ||
		strings.Contains(low, "already")
	if !noopish {
		t.Fatalf("dry-run current should claim noop/up to date/current; got:\n%s", resp.Stdout)
	}
	// Plan must not present InstallCmd as the action for a current install.
	if strings.Contains(resp.Stdout, codexinstall.InstallCmd) {
		t.Fatalf("dry-run current must not plan InstallCmd; got:\n%s", resp.Stdout)
	}
	// Avoid "would update" wording.
	if strings.Contains(low, "would update") {
		t.Fatalf("dry-run current must not plan update; got:\n%s", resp.Stdout)
	}
}
```
