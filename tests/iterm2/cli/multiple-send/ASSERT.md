## Expected

- `grok` before `codex` in captured script.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%s", resp.ExitCode, resp.Stderr)
	}
	s := resp.CapturedScript
	iGrok := strings.Index(s, `write text "grok"`)
	iCodex := strings.Index(s, `write text "codex"`)
	if iGrok < 0 || iCodex < 0 {
		t.Fatalf("script=%q", s)
	}
	if iGrok > iCodex {
		t.Fatal("order must be grok then codex")
	}
}
```