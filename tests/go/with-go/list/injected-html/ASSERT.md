## Expected

- Stdout is `go1.22.1` then `go1.19.13` (`go` + naked), document order.
- FetchHTML is called (injected hook used — not `go run download-go`).

## Expected Output

```
---
version: 3
---
go1.22.1
go1.19.13
```

## Errors

- `err` and `resp.Err` are nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("ListWith failed: %v", resp.Err)
	}
	if resp.FetchCount < 1 {
		t.Fatal("FetchHTML not called; ListWith must use downloadgo.List, not go run download-go")
	}
	assert.Output(t, resp.Stdout, `---
version: 3
---
go1.22.1
go1.19.13
`)
}
```
