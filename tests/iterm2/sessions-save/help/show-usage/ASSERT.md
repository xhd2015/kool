## Expected

- Exit 0
- Help mentions `save`, `restore`, `snapshot`

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
		t.Fatalf("exit=%d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	for _, want := range []string{"save", "restore", "snapshot"} {
		if !strings.Contains(resp.Stdout, want) {
			t.Fatalf("help missing %q:\n%s", want, resp.Stdout)
		}
	}
}
```
