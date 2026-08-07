## Expected Output

```
content-B-from-pack
```

## Expected

- Build succeeds; sealed run exit 0.
- Stdout equals packed content B (seed content A must not appear).

## Exit Code

- sealed run: 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("build exit=%d want 0; stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if !resp.RunExecuted {
		t.Fatal("expected sealed binary run")
	}
	if resp.RunExitCode != 0 {
		t.Fatalf("sealed exit=%d want 0; stdout=%q stderr=%q", resp.RunExitCode, resp.RunStdout, resp.RunStderr)
	}
	if strings.Contains(resp.RunStdout, "content-A-from-real-home") {
		t.Fatalf("guest must see packed B, not seed A; stdout=%q", resp.RunStdout)
	}
	assert.Output(t, resp.RunStdout, `---
version: 3
---
content-B-from-pack
`)
}
```
