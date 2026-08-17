## Expected

- Exit 0 (multi-match is success for find).
- Both `Alpha One` and `Alpha Two` listed; Beta preferably absent.

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
		t.Fatalf("find multi exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	if !strings.Contains(out, "Alpha One") || !strings.Contains(out, "Alpha Two") {
		t.Fatalf("find multi should list both Alpha tasks; out:\n%s", out)
	}
}
```
