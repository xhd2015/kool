## Expected

- Exit code 0.
- Stdout mentions `tab-set`, `list`, `run`, and config path or `tab-set` config dir
  (e.g. `config/iterm2/tab-set` or `KOOL_ITERM2_TAB_SET_DIR` / `.config`).

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := strings.ToLower(resp.Stdout + resp.Stderr)
	// Must be tab-set help, not open-dir usage only.
	if !strings.Contains(out, "tab-set") {
		t.Fatalf("help missing tab-set; out:\n%s", resp.Stdout)
	}
	for _, want := range []string{"list", "run", "show"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help missing %q; out:\n%s", want, resp.Stdout)
		}
	}
	if !strings.Contains(out, "config") && !strings.Contains(out, ".config") {
		t.Fatalf("help should mention config path; out:\n%s", resp.Stdout)
	}
}
```

