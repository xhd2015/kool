## Expected

- Exit 0.
- `toolchainVersions` is 2.
- toolchain versions have `legacy: false`.
- `legacyVersions` is 0.

## Exit Code

- 0

```go
import (
	"encoding/json"
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
	if rep.ToolchainVers != 2 {
		t.Fatalf("toolchainVersions=%d want 2", rep.ToolchainVers)
	}
	if rep.LegacyVersions != 0 {
		t.Fatalf("legacyVersions=%d want 0 (toolchain excluded)", rep.LegacyVersions)
	}
	if rep.SaveBytes != 0 {
		t.Fatalf("saveBytes=%d want 0 (toolchain excluded)", rep.SaveBytes)
	}
	for _, m := range rep.Modules {
		if m.Path != "golang.org/toolchain" {
			continue
		}
		for _, v := range m.Versions {
			if v.Legacy {
				t.Fatalf("toolchain %s marked legacy", v.Version)
			}
		}
	}
}
```
