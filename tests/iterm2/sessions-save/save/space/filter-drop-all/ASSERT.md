## Expected

- Exit 0
- Stdout contains 0 critical
- No file written
- Stderr skip warning for 2 windows

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
	if !strings.Contains(resp.Stdout, "0 critical") {
		t.Fatalf("stdout:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stderr, "skipped 2 windows not matching --spaces 1") {
		t.Fatalf("stderr:\n%s", resp.Stderr)
	}
	p := filepath.Join(req.WorkingDir, "filter-drop-all.json")
	if _, e := os.Stat(p); !os.IsNotExist(e) {
		t.Fatalf("must not write %s", p)
	}
}
```
