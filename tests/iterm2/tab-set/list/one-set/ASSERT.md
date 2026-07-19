## Expected

- Exit 0.
- Stdout contains set name `bots`.
- Prefer also mentioning tab count `2` (two tabs in fixture).

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout + resp.Stderr
	if !strings.Contains(out, "bots") {
		t.Fatalf("list missing bots; out:\n%s", out)
	}
	// Optional but preferred: tab count
	if !strings.Contains(out, "2") && !strings.Contains(strings.ToLower(out), "tab") {
		// require at least bots; soft prefer tabs
		t.Logf("list output has bots but no tab count hint: %s", out)
	}
}
```
