## Expected

- Exit code 0.
- Stdout help text mentions **`--no-enrich`**.
- Stdout mentions a **tree-related flag**: preferably `--no-tree`, or otherwise
  clear wording that documents suppressing the process tree (must still contain
  the substring `--no-tree` **or** both `tree` and a no-/disable style flag —
  locked: require `--no-tree` for implementer contract).
- Stdout ends with a trailing newline.

## Errors

- None.

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
	out := resp.Stdout
	for _, want := range []string{"--no-enrich", "--no-tree"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help missing %q:\n%s", want, out)
		}
	}
	if !strings.HasSuffix(out, "\n") {
		t.Fatalf("help stdout must end with newline; got %q", out)
	}
}
```
