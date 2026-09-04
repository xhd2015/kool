## Expected

- Exit 0.
- Stderr has `[1/3] extracted` (walking or sizing) and `ok`, plus download and vcs stages.
- Stdout has `GOMODCACHE:` and does not contain `[1/3]`.

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
	if !strings.Contains(resp.Stdout, "GOMODCACHE:") {
		t.Fatalf("expected report on stdout; got:\n%s", resp.Stdout)
	}
	if strings.Contains(resp.Stdout, "[1/3]") {
		t.Fatalf("stage markers must not be on stdout; got:\n%s", resp.Stdout)
	}
	errOut := resp.Stderr
	if !strings.Contains(errOut, "[1/3] extracted") {
		t.Fatalf("expected extracted stage on stderr; got:\n%s", errOut)
	}
	if !strings.Contains(errOut, "ok") {
		t.Fatalf("expected extracted ok on stderr; got:\n%s", errOut)
	}
	if !strings.Contains(errOut, "[2/3] download") {
		t.Fatalf("expected download stage; got:\n%s", errOut)
	}
	if !strings.Contains(errOut, "[3/3] vcs") {
		t.Fatalf("expected vcs stage; got:\n%s", errOut)
	}
	if !strings.Contains(errOut, "sizing 2 versions") {
		t.Fatalf("expected sizing count; got:\n%s", errOut)
	}
}
```
