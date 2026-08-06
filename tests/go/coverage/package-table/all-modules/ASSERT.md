## Expected

- Exit 0.
- Both packages present under `--all`.
- Package path for foreign file is directory of full cover path:
  `other.com/lib` (and module-local `example.com/mod/cli` or short `cli` —
  **locked:** with `--all`, package path is the directory of the full file path,
  so expect `` `example.com/mod/cli` `` and `` `other.com/lib` ``).
- Sorted by coverage ascending: other.com/lib 0.0% then example.com/mod/cli 100.0%.
- Stdout ends with `\n`.

## Expected Output

```
| Coverage | Package |
|----------|---------|
| 0.0% | `other.com/lib` |
| 100.0% | `example.com/mod/cli` |
```

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
	assert.Output(t, resp.Stdout, "---\n"+
		"version: 3\n"+
		"---\n"+
		`\| Coverage \| Package \|`+"\n"+
		`\|----------\|---------\|`+"\n"+
		`\| 0\.0% \| `+"`other.com/lib`"+` \|`+"\n"+
		`\| 100\.0% \| `+"`example.com/mod/cli`"+` \|`+"\n")
}
```
