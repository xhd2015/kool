## Expected Output

```
| Coverage | Package |
|----------|---------|
```

## Expected

- Exit **0**.
- Stdout is header-only table (no data rows), trailing `\n`.
- Stderr contains `warning:` (case-insensitive) about no packages / empty match.

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
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assert.Output(t, resp.Stdout, "---\n"+
		"version: 3\n"+
		"---\n"+
		`\| Coverage \| Package \|`+"\n"+
		`\|----------\|---------\|`+"\n")
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "warning:") {
		t.Fatalf("stderr must contain warning:; got %q", resp.Stderr)
	}
}
```
