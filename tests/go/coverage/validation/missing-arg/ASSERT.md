## Expected

- Exit code **2** (usage).
- Stderr non-empty (usage or required profile/coverage.out).

## Errors

- Missing required profile argument.

## Exit Code

- 2

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
	if resp.ExitCode != 2 {
		t.Fatalf("exit=%d want 2 for usage; stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatal("expected usage/error on stderr when profile arg missing")
	}
}
```
