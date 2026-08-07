## Expected

- Build succeeds; sealed run exit 0.
- First stdout line (`HOME`) equals second (`SANDBOX_ROOT`).
- That path is absolute, a session child under `SandboxRootParent`, and **not**
  the injected fake real home (`SealedHome`).

## Exit Code

- sealed run: 0

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("build exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !resp.RunExecuted {
		t.Fatal("expected sealed binary run")
	}
	if resp.RunExitCode != 0 {
		t.Fatalf("sealed exit=%d want 0; stdout=%q stderr=%q", resp.RunExitCode, resp.RunStdout, resp.RunStderr)
	}
	lines := strings.Split(strings.TrimSpace(resp.RunStdout), "\n")
	if len(lines) < 2 {
		t.Fatalf("want at least 2 lines (HOME, SANDBOX_ROOT); got %q", resp.RunStdout)
	}
	home := strings.TrimSpace(lines[0])
	sandboxRoot := strings.TrimSpace(lines[1])
	if home == "" || sandboxRoot == "" {
		t.Fatalf("empty HOME or SANDBOX_ROOT; stdout=%q", resp.RunStdout)
	}
	if home != sandboxRoot {
		t.Fatalf("HOME %q != SANDBOX_ROOT %q", home, sandboxRoot)
	}
	if !filepath.IsAbs(home) {
		t.Fatalf("HOME should be absolute; got %q", home)
	}
	parent := resp.SandboxRootParent
	if parent == "" {
		t.Fatal("expected SandboxRootParent recorded")
	}
	rel, err := filepath.Rel(parent, home)
	if err != nil {
		t.Fatalf("Rel(%q, %q): %v", parent, home, err)
	}
	if rel == "." || rel == "" {
		t.Fatalf("HOME should be a session child under parent, not the parent itself; HOME=%q parent=%q", home, parent)
	}
	if strings.HasPrefix(rel, "..") {
		t.Fatalf("HOME %q is not under KOOL_SANDBOX_ROOT %q (rel=%q)", home, parent, rel)
	}
	if resp.SealedHome != "" && home == resp.SealedHome {
		t.Fatalf("guest HOME must not remain the fake real home %q", resp.SealedHome)
	}
}
```
