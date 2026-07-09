## Expected Output

```
fail-once
```

## Expected

- Non-zero exit after the first failure.
- Exactly one child stdout line (did not continue to max-runs 5).

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
		t.Fatalf("expected non-zero on first failure with --allow-failure; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assert.Output(t, resp.Stdout, `---
version: 2
---
fail-once
`)
}
```
