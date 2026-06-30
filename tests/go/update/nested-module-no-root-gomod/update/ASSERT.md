## Expected

- `Update` returns no error
- Consumer go.mod requires `github.com/example/dot-pkgs/go-pkgs@v0.0.2`
- Replace directive for the module is dropped

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error running test: %v", err)
	}
	if resp.UpdateErr != nil {
		t.Fatalf("Update(%q) failed: %v", req.TargetDir, resp.UpdateErr)
	}
	if resp.ModuleVersion != "v0.0.2" {
		t.Fatalf("require version = %q, want v0.0.2", resp.ModuleVersion)
	}

	modInfo, err := resolve.GetModuleInfo(req.ConsumerDir)
	if err != nil {
		t.Fatalf("failed to read consumer go.mod: %v", err)
	}
	for _, repl := range modInfo.Replace {
		if repl.Old.Path == "github.com/example/dot-pkgs/go-pkgs" {
			t.Fatalf("replace for go-pkgs was not dropped: %+v", repl)
		}
	}
}
```