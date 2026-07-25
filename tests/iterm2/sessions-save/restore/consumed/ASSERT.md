## Expected

- Non-zero exit
- Error mentions consumed / restored_at

## Exit Code

- ≠ 0

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
	if resp.ExitCode == 0 {
		t.Fatal("expected consumed error")
	}
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "consumed") && !strings.Contains(low, "restored_at") {
		t.Fatalf("stderr:\n%s", resp.Stderr)
	}
}
```
