## Expected Output

```
. some.com/root
```

## Expected

- Exit code 0, no stderr.
- stdout contains exactly one line: `. some.com/root`.
- No line whose first field is `ext` or starts with `ext/` — the nested separate repo is
  skipped by the scan package.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error running test: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout: %s\nstderr: %s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}

	lines := stdoutLines(resp.Stdout)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d: %v\nfull stdout:\n%s", len(lines), lines, resp.Stdout)
	}
	if lines[0] != ". some.com/root" {
		t.Fatalf("only line = %q, want \". some.com/root\"", lines[0])
	}
	for _, line := range lines {
		dir := strings.SplitN(line, " ", 2)[0]
		if dir == "ext" || strings.HasPrefix(dir, "ext/") {
			t.Fatalf("nested separate repo line should be absent: %q", line)
		}
	}
}
```
