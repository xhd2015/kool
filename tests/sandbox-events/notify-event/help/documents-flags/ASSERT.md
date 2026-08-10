## Expected

- Exit 0.
- Stdout documents flag names exactly `--type` and `--path`.
- Stdout mentions notify-event or event (case-insensitive).
- Stdout ends with trailing `\n`.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		// RED until notify-event subcommand exists (unrecognized command → non-zero).
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stdout == "" {
		t.Fatal("expected help on stdout")
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("help stdout must end with newline; got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `<contains>
--type
--path
</contains>
`)
	low := strings.ToLower(resp.Stdout)
	if !strings.Contains(low, "notify") && !strings.Contains(low, "event") {
		t.Fatalf("help should mention notify/event; got:\n%s", resp.Stdout)
	}
}
```
