## Expected

- Exit 0.
- Stdout contains `removed`.
- example.com/foo@v1.0.0 extracted dir and zip are gone.
- example.com/foo@v1.2.0 remains.
- Both golang.org/toolchain versions remain.

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"golang.org/x/mod/module"
)

func extracted(t *testing.T, modcache, path, ver string) string {
	t.Helper()
	escaped, err := module.EscapePath(path)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(modcache, filepath.FromSlash(escaped)+"@"+ver)
}

func zipPath(t *testing.T, modcache, path, ver string) string {
	t.Helper()
	escaped, err := module.EscapePath(path)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(modcache, "cache", "download", filepath.FromSlash(escaped), "@v", ver+".zip")
}

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "removed") {
		t.Fatalf("apply should say removed; got:\n%s", resp.Stdout)
	}
	if strings.Contains(resp.Stdout, "would remove") {
		t.Fatalf("apply must not say would remove; got:\n%s", resp.Stdout)
	}
	oldDir := extracted(t, req.ModCache, "example.com/foo", "v1.0.0")
	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Fatalf("legacy extracted dir should be gone: %s err=%v", oldDir, err)
	}
	oldZip := zipPath(t, req.ModCache, "example.com/foo", "v1.0.0")
	if _, err := os.Stat(oldZip); !os.IsNotExist(err) {
		t.Fatalf("legacy zip should be gone: %s err=%v", oldZip, err)
	}
	newDir := extracted(t, req.ModCache, "example.com/foo", "v1.2.0")
	if _, err := os.Stat(newDir); err != nil {
		t.Fatalf("newest extracted dir should remain: %s: %v", newDir, err)
	}
	newZip := zipPath(t, req.ModCache, "example.com/foo", "v1.2.0")
	if _, err := os.Stat(newZip); err != nil {
		t.Fatalf("newest zip should remain: %s: %v", newZip, err)
	}
	for _, ver := range []string{"v0.0.1-go1.21.0.darwin-arm64", "v0.0.1-go1.22.0.darwin-arm64"} {
		p := extracted(t, req.ModCache, "golang.org/toolchain", ver)
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("toolchain %s should remain: %v", p, err)
		}
	}
}
```
