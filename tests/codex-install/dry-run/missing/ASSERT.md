## Expected

- Exit 0.
- Stdout mentions dry-run and an install plan (would install / install action).
- Stdout includes `install.InstallCmd` (exact library constant), or at least the
  install script URL / `curl` install recipe.
- `ShellCalls` empty (no mutation).
- `FetchLatestCalls == 0` (missing path must not require latest).

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
	if resp.FetchLatestCalls != 0 {
		t.Fatalf("FetchLatestCalls = %d, want 0 when bin missing", resp.FetchLatestCalls)
	}
	assert.Output(t, resp.Stdout, `<contains>
dry-run
</contains>
`)
	low := strings.ToLower(resp.Stdout)
	if !strings.Contains(low, "install") {
		t.Fatalf("dry-run missing plan should mention install; got:\n%s", resp.Stdout)
	}
	// Prefer exact InstallCmd; accept URL fragment if implementer prints URL only.
	if !strings.Contains(resp.Stdout, codexinstall.InstallCmd) &&
		!strings.Contains(resp.Stdout, codexinstall.InstallScriptURL) &&
		!strings.Contains(resp.Stdout, "codex/install.sh") {
		t.Fatalf("dry-run missing plan should include InstallCmd or install script URL; got:\n%s\nwant InstallCmd=%q",
			resp.Stdout, codexinstall.InstallCmd)
	}
}
```
