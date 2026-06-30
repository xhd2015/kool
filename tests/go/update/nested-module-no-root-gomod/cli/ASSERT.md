## Expected

- `kool go update` exits with code 0
- Consumer go.mod requires `github.com/example/dot-pkgs/go-pkgs@v0.0.2`
- Replace directive for the module is dropped

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error running test: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout: %s\nstderr: %s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if resp.UpdateErr != nil {
		t.Fatalf("kool go update failed: %v\nstdout: %s\nstderr: %s", resp.UpdateErr, resp.Stdout, resp.Stderr)
	}
	if resp.ModuleVersion != "v0.0.2" {
		t.Fatalf("require version = %q, want v0.0.2\nstdout: %s\nstderr: %s", resp.ModuleVersion, resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "commit message:") {
		t.Fatalf("expected commit message in stdout, got:\n%s", resp.Stdout)
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