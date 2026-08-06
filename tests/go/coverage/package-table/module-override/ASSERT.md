## Expected Output

```
| Coverage | Package |
|----------|---------|
| 100.0% | `svc` |
```

## Expected

- Exit 0.
- Only packages under `--module example.com/other` (`svc`).
- `cli` from example.com/mod omitted despite go.mod.

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
		`\| 100\.0% \| `+"`svc`"+` \|`+"\n")
	if strings.Contains(resp.Stdout, "`cli`") {
		t.Fatalf("--module should exclude example.com/mod/cli; got:\n%s", resp.Stdout)
	}
}
```
