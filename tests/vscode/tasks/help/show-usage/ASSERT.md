## Expected

- Exit code 0.
- Output mentions `tasks`, and subcommands `list`, `find`, `show`, `run`.
- Prefer mentioning `--dry-run` and/or `tasks.json` / workspace.

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
		t.Fatalf("help exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := strings.ToLower(resp.Stdout + resp.Stderr)
	if !strings.Contains(out, "tasks") {
		t.Fatalf("help missing tasks; out:\n%s", resp.Stdout)
	}
	for _, want := range []string{"list", "find", "show", "run"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help missing %q; out:\n%s", want, resp.Stdout)
		}
	}
	if !strings.Contains(out, "dry-run") && !strings.Contains(out, "dry run") {
		t.Logf("help lacks --dry-run mention (prefer): %s", resp.Stdout)
	}
}
```
