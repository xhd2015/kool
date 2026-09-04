## Expected

- Exit 0.
- JSON modules include `github.com/Azure/ok` (decoded), not `!azure`.

## Exit Code

- 0

```go
import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/kool/tools/go/modcache"
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
	var rep modcache.Report
	if err := json.Unmarshal([]byte(resp.Stdout), &rep); err != nil {
		t.Fatalf("json: %v\n%s", err, resp.Stdout)
	}
	found := false
	for _, m := range rep.Modules {
		if m.Path == "github.com/Azure/ok" {
			found = true
		}
		if strings.Contains(m.Path, "!azure") {
			t.Fatalf("path should be unescaped; got %q", m.Path)
		}
	}
	if !found {
		t.Fatalf("expected github.com/Azure/ok; got %+v", rep.Modules)
	}
}
```
