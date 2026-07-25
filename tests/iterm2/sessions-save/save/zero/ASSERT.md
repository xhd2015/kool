## Expected

- Exit 0
- Message about 0 critical
- empty.json not created

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
		t.Fatalf("exit=%d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "0 critical") {
		t.Fatalf("stdout:\n%s", resp.Stdout)
	}
	p := filepath.Join(req.WorkingDir, "empty.json")
	if _, e := os.Stat(p); !os.IsNotExist(e) {
		t.Fatalf("must not write %s", p)
	}
}
```
