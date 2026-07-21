## Expected

- Build succeeds; sealed run exit 0.
- First stdout line (`SANDBOX_ROOT`) equals second line (`pwd`).
- That path is absolute and a session child under `SandboxRootParent`.

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
	lines := strings.Split(strings.TrimSpace(resp.RunStdout), "\n")
	if len(lines) < 2 {
		t.Fatalf("want at least 2 lines (SANDBOX_ROOT, pwd); got %q", resp.RunStdout)
	}
	sandboxRoot := strings.TrimSpace(lines[0])
	pwd := strings.TrimSpace(lines[1])
	if sandboxRoot == "" || pwd == "" {
		t.Fatalf("empty SANDBOX_ROOT or pwd; stdout=%q", resp.RunStdout)
	}
	if sandboxRoot != pwd {
		t.Fatalf("SANDBOX_ROOT %q != cwd/pwd %q", sandboxRoot, pwd)
	}
	if !filepath.IsAbs(sandboxRoot) {
		t.Fatalf("SANDBOX_ROOT should be absolute; got %q", sandboxRoot)
	}
	parent := resp.SandboxRootParent
	rel, err := filepath.Rel(parent, sandboxRoot)
	if err != nil {
		t.Fatalf("Rel: %v", err)
	}
	if rel == "." || strings.HasPrefix(rel, "..") {
		t.Fatalf("SANDBOX_ROOT %q not a session child of %q (rel=%q)", sandboxRoot, parent, rel)
	}
}
```
