## Expected

- Exit 0
- Help mentions auto-backup, --interval, 10m, --once, sessions-auto.json, --file, --dry-run

## Exit Code

- 0

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
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout + "\n" + resp.Stderr
	for _, want := range []string{
		"auto-backup",
		"--interval",
		"10m",
		"--once",
		"sessions-auto.json",
		"--file",
		"--dry-run",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("auto-backup help missing %q:\n%s", want, out)
		}
	}
}
```
