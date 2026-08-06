## Expected Output

```
| Coverage | Package |
|----------|---------|
| 0.0% | `internal/run` |
| 100.0% | `cli` |
```

Output ends with a trailing newline.

## Expected

- Exit 0.
- Exact markdown table sorted by coverage ascending, then package name.
- Stdout ends with `\n`.

## Exit Code

- 0

```go
import (
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
	// v3 lines are raw Go regexps — escape | and .
	assert.Output(t, resp.Stdout, "---\n"+
		"version: 3\n"+
		"---\n"+
		`\| Coverage \| Package \|`+"\n"+
		`\|----------\|---------\|`+"\n"+
		`\| 0\.0% \| `+"`internal/run`"+` \|`+"\n"+
		`\| 100\.0% \| `+"`cli`"+` \|`+"\n")
}
```
