## Expected Output

```
fail-line
fail-line
```

## Expected

- Non-zero exit after two consecutive failures.
- Exactly two child stdout lines (max-failure trips before max-runs).

## Exit Code

- non-zero

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero after max-failure trip; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assert.Output(t, resp.Stdout, `---
version: 2
---
fail-line
fail-line
`)
}
```
