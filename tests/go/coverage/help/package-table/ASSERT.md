## Expected

- Exit 0.
- Stdout documents principal flags: `--module`, `--dir`, `--skip-prefix`,
  `--skip-contains`, `--all`, `--json` (and profile / coverage.out).
- Stdout ends with a trailing newline.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
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
	if resp.Stdout == "" {
		t.Fatal("expected package-table help on stdout")
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("help stdout must end with newline; got %q", resp.Stdout)
	}
	for _, want := range []string{
		"--module",
		"--dir",
		"--skip-prefix",
		"--skip-contains",
		"--all",
		"--json",
	} {
		if !strings.Contains(resp.Stdout, want) {
			t.Fatalf("package-table help missing %q:\n%s", want, resp.Stdout)
		}
	}
}
```
