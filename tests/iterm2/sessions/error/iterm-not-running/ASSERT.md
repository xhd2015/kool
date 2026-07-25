## Expected

- Exit code non-zero (typically 1).
- Stderr mentions iTerm2 and/or “not running” (case-insensitive).
- Stdout empty or without a successful full snapshot header of success path.

## Errors

- Reported on stderr; Run returns exit code via RunForTest (err nil).

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
		t.Fatalf("want non-zero exit when iTerm not running; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	msg := strings.ToLower(resp.Stderr + " " + resp.Stdout)
	if !strings.Contains(msg, "iterm") && !strings.Contains(msg, "not running") {
		t.Fatalf("stderr/stdout should mention iTerm / not running: %q", resp.Stderr)
	}
}
```
