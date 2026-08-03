## Expected

- Exit 0
- Checkpoint written (Saved)
- FileJSON does **not** contain `"space"` or `"iterm_window_id"` keys

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
	if resp.FileJSON == "" {
		t.Fatal("expected checkpoint file written")
	}
	// Omit both Space fields when --ignore-macos-space.
	if strings.Contains(resp.FileJSON, `"space"`) {
		t.Fatalf("--ignore-macos-space must omit \"space\":\n%s", resp.FileJSON)
	}
	if strings.Contains(resp.FileJSON, `"iterm_window_id"`) {
		t.Fatalf("--ignore-macos-space must omit \"iterm_window_id\":\n%s", resp.FileJSON)
	}
}
```
