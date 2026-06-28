## Expected
- Non-zero exit code.
- Stderr mentions `code` or PATH.

## Exit Code
- 1

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error running kool: %v", err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit when code missing, got 0\nstderr: %s", resp.Stderr)
	}
	assert.Output(t, resp.Stderr, `
<contains>
code
</contains>`)
}
```