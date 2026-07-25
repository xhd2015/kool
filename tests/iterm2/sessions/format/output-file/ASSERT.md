## Expected

- Exit 0.
- Stdout empty (or whitespace-only).
- Stderr contains `Wrote`.
- Output file non-empty; contains `"source"` / `iterm2` or fixture session id.
- Buffered path (`SawW1BeforeLastListTabs` false).

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
	if strings.TrimSpace(resp.Stdout) != "" {
		t.Fatalf("stdout should be empty when -o set; got %q", resp.Stdout)
	}
	if !strings.Contains(resp.Stderr, "Wrote") {
		t.Fatalf("stderr missing Wrote: %q", resp.Stderr)
	}
	if strings.TrimSpace(resp.OutputFile) == "" {
		t.Fatal("output file empty or unreadable")
	}
	if !strings.Contains(resp.OutputFile, "iterm2") && !strings.Contains(resp.OutputFile, "AAAAAAAA") {
		t.Fatalf("output file missing expected content:\n%s", resp.OutputFile)
	}
	if resp.SawW1BeforeLastListTabs {
		t.Fatal("-o must buffer (no progressive CLI during collect)")
	}
}
```
