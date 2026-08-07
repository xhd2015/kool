## Expected Output

```
content-C-packed
---
content-O-sibling
```

## Expected

- Build succeeds; sealed run exit 0.
- Packed path `.config/mytool/cfg` shows content C.
- Sibling seed path `.config/other/x` still shows content O after explode.

## Exit Code

- sealed run: 0

```go
import (
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
	assert.Output(t, resp.RunStdout, `---
version: 3
---
content-C-packed
---
content-O-sibling
`)
}
```
