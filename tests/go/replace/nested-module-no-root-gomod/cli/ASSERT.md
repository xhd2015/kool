## Expected

- `kool go replace` exits with code 0
- Consumer go.mod contains `replace github.com/example/dot-pkgs/go-pkgs => <abs go-pkgs dir>`

```go
import (
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
	if resp.ReplaceErr != nil {
		t.Fatalf("kool go replace failed: %v\nstdout: %s\nstderr: %s", resp.ReplaceErr, resp.Stdout, resp.Stderr)
	}
	if resp.ModulePath != "github.com/example/dot-pkgs/go-pkgs" {
		t.Fatalf("module path = %q, want github.com/example/dot-pkgs/go-pkgs", resp.ModulePath)
	}
	if !resp.HasReplace {
		t.Fatalf("expected replace directive for %s -> %s", resp.ModulePath, resp.AbsDir)
	}

	modInfo, err := resolve.GetModuleInfo(req.ConsumerDir)
	if err != nil {
		t.Fatalf("failed to read consumer go.mod: %v", err)
	}
	found := false
	for _, repl := range modInfo.Replace {
		if repl.Old.Path == "github.com/example/dot-pkgs/go-pkgs" && repl.New.Path == resp.AbsDir {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("replace for go-pkgs not found in go.mod: %+v", modInfo.Replace)
	}
}
```