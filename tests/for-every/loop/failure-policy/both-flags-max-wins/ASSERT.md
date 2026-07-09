## Expected Output

```
fail-line
fail-line
fail-line
```

## Expected

- Non-zero exit after **3** consecutive failures (max-failure wins).
- Not 1 line (which would mean allow-failure alone applied).
- Not 10 lines (max-runs safety only).

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
		t.Fatalf("expected non-zero after max-failure 3; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
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
