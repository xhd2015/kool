## Expected

- Exit 1 with a validation error about the missing title (not open-dir path errors).
- Stderr should mention title (or usage including set-title), and must not treat
  `set-title` as a directory path (`stat` of set-title).
- Must not print a success `title changed:` line.

## Errors

- Validation error before a successful title change.

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
		t.Fatal("expected validation error on stderr")
	}
	// Must be title validation, not open-dir treating "set-title" as a path.
	low := strings.ToLower(resp.Stderr)
	if strings.Contains(low, "stat") || strings.Contains(low, "no such file") {
		t.Fatalf("set-title must be a reserved subcommand, not a directory path:\nstderr=%q", resp.Stderr)
	}
	if !strings.Contains(low, "title") && !strings.Contains(low, "set-title") && !strings.Contains(low, "usage") {
		t.Fatalf("expected title/usage validation message, stderr=%q", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "title changed:") {
		t.Fatalf("unexpected success stdout: %q", resp.Stdout)
	}
}
```
