## Expected Output

```
| Coverage | Package |
|----------|---------|
| 0.0% | `pkg/keep` |
| 100.0% | `script/gen` |
```

## Expected

- Exit 0.
- Custom skip lists **replace** defaults: `script/gen` is kept; `tmp/x` and
  `pkg/x_drop` omitted.
- Sorted by coverage then name.
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
		`\| 0\.0% \| `+"`pkg/keep`"+` \|`+"\n"+
		`\| 100\.0% \| `+"`script/gen`"+` \|`+"\n")
	for _, banned := range []string{"`tmp/x`", "x_drop"} {
		if strings.Contains(resp.Stdout, banned) {
			t.Fatalf("custom skips should omit %q; got:\n%s", banned, resp.Stdout)
		}
	}
}
```
