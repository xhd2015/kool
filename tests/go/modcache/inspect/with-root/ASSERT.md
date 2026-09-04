## Expected

- Exit 0.
- JSON suggestions include example.com/foo current v1.0.0 newest v1.2.0.

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
	if !strings.Contains(resp.Stderr, "[1/4]") {
		t.Fatalf("with --root, stages should be 1/4; stderr:\n%s", resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "[4/4] live") {
		t.Fatalf("expected live stage; stderr:\n%s", resp.Stderr)
	}
	var rep modcache.Report
	if err := json.Unmarshal([]byte(resp.Stdout), &rep); err != nil {
		t.Fatalf("json: %v\n%s", err, resp.Stdout)
	}
	if rep.GoModules < 1 {
		t.Fatalf("goModules=%d want >=1; stderr=%q", rep.GoModules, resp.Stderr)
	}
	found := false
	for _, s := range rep.Suggestions {
		if s.Path == "example.com/foo" && s.Current == "v1.0.0" && s.Newest == "v1.2.0" {
			found = true
			if len(s.Repos) == 0 {
				t.Fatal("suggestion should list consuming module dirs")
			}
		}
	}
	if !found {
		t.Fatalf("expected upgrade suggestion foo v1.0.0 -> v1.2.0; got %+v", rep.Suggestions)
	}
}
```
