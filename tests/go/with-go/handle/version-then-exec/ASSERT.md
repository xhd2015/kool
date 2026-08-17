## Expected

- `HandleWith(["go1.19", "go"])` succeeds.
- Install hook is unused (dest already exists).
- Child GOROOT is `filepath.Abs($InstallDir/go1.19.13)`.
- First PATH entry is that dest's `bin`.

## Expected Output

```
---
version: 3
__GOROOT__: type=string, example=/tmp/installed/go1.19.13
__BIN__: type=string, example=/tmp/installed/go1.19.13/bin
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
		t.Fatalf("HandleWith(%v) failed: %v", req.Args, resp.Err)
	}
	if resp.HookCalled {
		t.Fatalf("Install hook called (version=%q dir=%q); dest already existed", resp.HookVersion, resp.HookDir)
	}
	abs := absPath(t, destPin(req.InstallDir))
	bin := filepath.Join(abs, "bin")
	assert.Output(t, resp.Stdout, `---
version: 3
---
GOROOT=`+regexp.QuoteMeta(abs)+`
PATH0=`+regexp.QuoteMeta(bin)+`
`)
}
```
