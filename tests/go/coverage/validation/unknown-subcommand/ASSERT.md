## Expected

- Non-zero exit.
- Stderr mentions unknown/unrecognized/invalid (or the bad name `nosuch`).

## Errors

- Unknown subcommand.

## Exit Code

- non-zero

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
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for unknown subcommand; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatal("expected stderr for unknown subcommand")
	}
	low := strings.ToLower(resp.Stderr)
	ok := strings.Contains(low, "unknown") ||
		strings.Contains(low, "unrecognized") ||
		strings.Contains(low, "invalid") ||
		strings.Contains(low, "nosuch")
	if !ok {
		t.Fatalf("stderr should indicate unknown command; got %q", resp.Stderr)
	}
}
```
