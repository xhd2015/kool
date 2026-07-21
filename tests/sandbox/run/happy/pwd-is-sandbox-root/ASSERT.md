## Expected

- Build succeeds; sealed binary executed.
- Sealed exit 0.
- Trimmed stdout is an absolute path under `SandboxRootParent` (session child).

## Exit Code

- sealed run: 0

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("build exit=%d want 0; stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if !resp.RunExecuted {
		t.Fatal("expected sealed binary run")
	}
	if resp.RunExitCode != 0 {
		t.Fatalf("sealed exit=%d want 0; stdout=%q stderr=%q", resp.RunExitCode, resp.RunStdout, resp.RunStderr)
	}
	got := strings.TrimSpace(resp.RunStdout)
	if got == "" {
		t.Fatal("expected pwd on stdout")
	}
	if !filepath.IsAbs(got) {
		t.Fatalf("pwd should be absolute; got %q", got)
	}
	parent := resp.SandboxRootParent
	if parent == "" {
		t.Fatal("expected SandboxRootParent recorded")
	}
	rel, err := filepath.Rel(parent, got)
	if err != nil {
		t.Fatalf("Rel(%q, %q): %v", parent, got, err)
	}
	if rel == "." || rel == "" {
		t.Fatalf("pwd should be a session child under parent, not the parent itself; pwd=%q parent=%q", got, parent)
	}
	if strings.HasPrefix(rel, "..") {
		t.Fatalf("pwd %q is not under KOOL_SANDBOX_ROOT %q (rel=%q)", got, parent, rel)
	}
}
```
