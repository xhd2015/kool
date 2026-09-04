## Expected

- Exit 0.
- JSON `legacyVersions` is 1.
- example.com/foo newest is v1.2.0; v1.0.0 is legacy.

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
	if strings.Contains(resp.Stdout, "[1/3]") {
		t.Fatalf("JSON stdout must not contain stage markers; got:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stderr, "[1/3] extracted") {
		t.Fatalf("stage markers belong on stderr; got:\n%s", resp.Stderr)
	}
	var rep modcache.Report
	if err := json.Unmarshal([]byte(resp.Stdout), &rep); err != nil {
		t.Fatalf("json: %v\n%s", err, resp.Stdout)
	}
	if rep.LegacyVersions != 1 {
		t.Fatalf("legacyVersions=%d want 1", rep.LegacyVersions)
	}
	if rep.SaveBytes != rep.LegacyBytes || rep.SaveBytes <= 0 {
		t.Fatalf("saveBytes=%d legacyBytes=%d want equal and >0", rep.SaveBytes, rep.LegacyBytes)
	}
	var foo *modcache.ModuleJSON
	for i := range rep.Modules {
		if rep.Modules[i].Path == "example.com/foo" {
			foo = &rep.Modules[i]
			break
		}
	}
	if foo == nil {
		t.Fatalf("missing example.com/foo in %+v", rep.Modules)
	}
	if foo.Newest != "v1.2.0" {
		t.Fatalf("newest=%q want v1.2.0", foo.Newest)
	}
	var sawOld, sawNew bool
	for _, v := range foo.Versions {
		if v.Version == "v1.0.0" {
			sawOld = true
			if !v.Legacy {
				t.Fatal("v1.0.0 should be legacy")
			}
		}
		if v.Version == "v1.2.0" {
			sawNew = true
			if v.Legacy {
				t.Fatal("v1.2.0 should not be legacy")
			}
		}
	}
	if !sawOld || !sawNew {
		t.Fatalf("expected both versions; got %+v", foo.Versions)
	}
}
```
