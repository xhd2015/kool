## Expected

- Non-zero exit.
- Stderr is non-empty and mentions `package-table` and/or `help` / `usage`.

## Errors

- No subcommand provided.

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
		t.Fatalf("expected non-zero for bare coverage; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatal("expected stderr message for bare coverage")
	}
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "package-table") && !strings.Contains(low, "help") && !strings.Contains(low, "usage") {
		t.Fatalf("stderr should hint package-table/help/usage; got %q", resp.Stderr)
	}
}
```
