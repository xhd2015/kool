## Expected

- Exit 0; script contains `write text "grok"`.

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
		t.Fatalf("exit=%d stderr=%s", resp.ExitCode, resp.Stderr)
	}
	if !strings.Contains(resp.CapturedScript, writeTextCmd("grok")) {
		t.Fatalf("script=%q", resp.CapturedScript)
	}
}
```