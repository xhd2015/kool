## Expected

- Non-zero exit
- Stderr or stdout mentions cannot be specified together (color flags)

## Errors

- Flag conflict for `--color` and `--no-color`

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
		t.Fatalf("expected conflict failure; stdout=%q", resp.Stdout)
	}
	msg := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	if !strings.Contains(msg, "cannot be specified together") {
		t.Fatalf("expected “cannot be specified together”; stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}
}
```
