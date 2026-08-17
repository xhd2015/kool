## Expected

- Exit 1; stderr mentions macOS / unsupported platform.

## Exit Code

- 1

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
	if resp.ExitCode == 0 {
		t.Fatal("expected non-zero on linux")
	}
	combined := strings.ToLower(resp.Stderr + resp.Stdout)
	if !strings.Contains(combined, "macos") && !strings.Contains(combined, "unsupported") {
		t.Fatalf("stderr=%q", resp.Stderr)
	}
}
```