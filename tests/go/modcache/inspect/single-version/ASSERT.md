## Expected

- Exit 0.
- Stdout mentions example.com/foo only if it appears in TOP; TOP legacy table should be absent or not list this path as reclaimable.
- LEGACY 0 versions.

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
		t.Fatalf("exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if strings.Contains(resp.Stdout, "TOP legacy") {
		t.Fatalf("single version should not have TOP legacy table; got:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "LEGACY:") {
		t.Fatalf("expected LEGACY line; got:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "SAVE:") {
		t.Fatalf("expected SAVE line; got:\n%s", resp.Stdout)
	}
	if strings.Contains(resp.Stdout, "% of total") {
		t.Fatalf("zero save should not include percent of total; got:\n%s", resp.Stdout)
	}
}
```
