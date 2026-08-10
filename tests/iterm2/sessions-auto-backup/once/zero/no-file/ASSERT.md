## Expected

- Exit 0
- Message about 0 critical
- empty-auto.json not created

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout + "\n" + resp.Stderr
	if !strings.Contains(out, "0 critical") {
		t.Fatalf("expected 0 critical message:\n%s", out)
	}
	p := filepath.Join(req.WorkingDir, "empty-auto.json")
	if _, e := os.Stat(p); !os.IsNotExist(e) {
		t.Fatalf("must not write %s", p)
	}
}
```
