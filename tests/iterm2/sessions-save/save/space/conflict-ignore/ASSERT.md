## Expected

- Non-zero exit
- Stderr mentions --spaces cannot be used with --ignore-macos-space

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
		t.Fatalf("expected failure; stdout=%q", resp.Stdout)
	}
	msg := resp.Stderr + "\n" + resp.Stdout
	if !strings.Contains(msg, "--spaces cannot be used with --ignore-macos-space") {
		t.Fatalf("stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}
}
```
