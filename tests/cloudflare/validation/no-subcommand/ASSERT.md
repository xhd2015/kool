## Expected

- Non-zero exit.
- Non-empty stderr suggesting subcommand / help / usage / require.
- StartSession not called.

## Errors

- Missing subcommand.

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
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for bare root; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatal("expected stderr validation message")
	}
	if resp.StartCalled {
		t.Fatal("StartSession must not run when subcommand missing")
	}
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "usage") && !strings.Contains(low, "help") &&
		!strings.Contains(low, "command") && !strings.Contains(low, "require") &&
		!strings.Contains(low, "serve") && !strings.Contains(low, "subcommand") {
		t.Fatalf("stderr should suggest subcommands/help; got %q", resp.Stderr)
	}
}
```
