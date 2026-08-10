## Expected

- Exit 0
- Stdout contains Saved
- FileJSON includes `"app": "/Applications/iTerm.app"` (or FixtureApp)
- Never `"app": "/Users/…"` home form
- Version 1; not consumed

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
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "Saved") {
		t.Fatalf("stdout:\n%s", resp.Stdout)
	}
	if resp.FileJSON == "" {
		t.Fatal("missing FileJSON")
	}
	want := req.FixtureApp
	if want == "" {
		want = fixtureAppSystem
	}
	if !fileJSONHasApp(resp.FileJSON, want) {
		t.Fatalf("checkpoint must emit \"app\": %q when known:\n%s", want, resp.FileJSON)
	}
	// Home form never expanded.
	if strings.Contains(resp.FileJSON, `"app": "/Users/`) {
		t.Fatalf("home app must use ~/ form, not /Users/…:\n%s", resp.FileJSON)
	}
	if resp.Doc == nil {
		t.Fatal("missing parsed Doc")
	}
	if resp.Doc.Version != 1 {
		t.Fatalf("version=%d", resp.Doc.Version)
	}
	if resp.Doc.IsConsumed() {
		t.Fatal("restored_at should be empty after save")
	}
}
```
