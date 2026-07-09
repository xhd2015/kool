## Expected

- Exit 1; empty title is invalid (subcommand recognized).
- Stderr mentions empty and/or title (validation), not open-dir `stat` / bare
  `unrecognized arguments` from treating `set-title` as a directory.
- No success stdout line.

## Errors

- Empty title rejected.

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
		t.Fatal("expected validation error on stderr for empty title")
	}
	low := strings.ToLower(resp.Stderr)
	if strings.Contains(low, "stat") || strings.Contains(low, "no such file") {
		t.Fatalf("set-title must be reserved, not open-dir:\nstderr=%q", resp.Stderr)
	}
	// Must describe title validation (not open-dir "unrecognized arguments: …").
	if !strings.Contains(low, "title") && !strings.Contains(low, "empty") {
		t.Fatalf("expected empty-title validation message, stderr=%q", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "title changed:") {
		t.Fatalf("unexpected success stdout: %q", resp.Stdout)
	}
}
```
