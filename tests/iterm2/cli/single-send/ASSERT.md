## Expected

- Exit 0; script contains `write text "grok"`.

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
	if !strings.Contains(resp.CapturedScript, `write text "grok"`) {
		t.Fatalf("script=%q", resp.CapturedScript)
	}
}
```