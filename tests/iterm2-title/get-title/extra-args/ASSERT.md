## Expected

- Exit 1 when extra positionals are present on get-title.
- Stderr identifies the `get-title` subcommand (not open-dir treating
  `get-title` as a directory path with generic unrecognized-args only).
- Stderr mentions the unexpected arg (`foo`) or clearly rejects extras.

## Errors

- Extra arguments rejected by get-title handler.

## Exit Code

- 1

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 1 {
		t.Fatalf("exit=%d want 1\nstdout=%q\nstderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatal("expected error message on stderr for extra args")
	}
	low := strings.ToLower(resp.Stderr)
	if strings.Contains(low, "stat") || strings.Contains(low, "no such file") {
		t.Fatalf("get-title must be a reserved subcommand, not a directory path:\nstderr=%q", resp.Stderr)
	}
	// Require get-title in the message so open-dir's
	// "unrecognized arguments: foo" does not false-pass before the feature exists.
	if !strings.Contains(low, "get-title") {
		t.Fatalf("error should mention get-title (subcommand routing), stderr=%q", resp.Stderr)
	}
}
```
