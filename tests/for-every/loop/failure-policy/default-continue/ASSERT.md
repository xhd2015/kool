## Expected Output

```
fail-line
fail-line
fail-line
```

## Expected

- Non-zero exit (last run failed).
- Exactly three child stdout lines — loop continued after failures rather than stopping at first.
- Stderr may log each failure (not asserted strictly).

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
		t.Fatalf("expected non-zero after failed last run; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assert.Output(t, resp.Stdout, `---
version: 2
---
fail-line
fail-line
fail-line
`)
}
```
