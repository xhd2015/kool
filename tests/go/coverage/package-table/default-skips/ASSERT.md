## Expected Output

```
| Coverage | Package |
|----------|---------|
| 100.0% | `pkg/keep` |
```

## Expected

- Exit 0.
- Only `pkg/keep` appears; `script/gen`, `cmd/tool`, and `pkg/legacy_old` absent.
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
		`\| 100\.0% \| `+"`pkg/keep`"+` \|`+"\n")
	for _, banned := range []string{"script/gen", "cmd/tool", "legacy_old", "`script/", "`cmd/"} {
		if strings.Contains(resp.Stdout, banned) {
			t.Fatalf("stdout must omit skipped package path containing %q; got:\n%s", banned, resp.Stdout)
		}
	}
}
```
