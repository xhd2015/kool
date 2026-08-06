## Expected Output

```
| Coverage | Package |
|----------|---------|
| 100.0% | `cli` |
```

## Expected

- Exit 0.
- Only module-local package `cli`.
- Foreign `other.com/lib` absent (and not shown as package path).
- Stdout ends with `\n`.

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
		`\|----------\|---------\|`+"\n"+
		`\| 100\.0% \| `+"`cli`"+` \|`+"\n")
	if strings.Contains(resp.Stdout, "other.com") || strings.Contains(resp.Stdout, "`lib`") {
		t.Fatalf("foreign module package must be omitted; got:\n%s", resp.Stdout)
	}
}
```
