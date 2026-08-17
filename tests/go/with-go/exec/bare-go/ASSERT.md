## Expected

- Child GOROOT is `filepath.Abs(goroot)`.
- First PATH entry is `$abs/bin`.

## Expected Output

```
---
version: 3
__GOROOT__: type=string, example=/tmp/goroot
__BIN__: type=string, example=/tmp/goroot/bin
---
GOROOT=__GOROOT__
PATH0=__BIN__
```

## Errors

- `err` and `resp.Err` are nil.

```go
import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("ExecGoroot(%q, %v) failed: %v", req.Goroot, req.Args, resp.Err)
	}
	abs := absPath(t, req.Goroot)
	bin := filepath.Join(abs, "bin")
	assert.Output(t, resp.Stdout, `---
version: 3
---
GOROOT=`+regexp.QuoteMeta(abs)+`
PATH0=`+regexp.QuoteMeta(bin)+`
`)
}
```
