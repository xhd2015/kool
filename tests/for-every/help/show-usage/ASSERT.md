## Expected

- Exit 0.
- Stdout documents both `for-every <duration>` and glued `for-every-<duration>` forms.
- Stdout mentions `--max-runs`, `--max-failure`, and `--allow-failure`.
- Lines end with newline convention for user-facing help (last content followed by `\n`).

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stdout == "" {
		t.Fatal("expected help on stdout")
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("help stdout must end with newline; got %q", resp.Stdout)
	}
	// Soft contains via v1 template (help text wording may vary slightly).
	// No leading blank line; trailing newline before closing backtick so
	// template agrees with CLI stdout ending in \n.
	assert.Output(t, resp.Stdout, `<contains>
for-every
--max-runs
--max-failure
--allow-failure
</contains>
`)
	low := strings.ToLower(resp.Stdout)
	if !strings.Contains(low, "duration") {
		t.Fatalf("help should mention duration; stdout=%q", resp.Stdout)
	}
}
```
