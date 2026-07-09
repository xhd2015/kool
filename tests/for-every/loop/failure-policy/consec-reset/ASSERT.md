## Expected Output

```
run-1
run-2
run-3
run-4
run-5
```

## Expected

- Completes all 5 max-runs despite three overall failures (runs 1,3,5).
- Consecutive failures never reach 2 because even runs succeed and reset the counter.
- If total (non-consecutive) failures were counted, the loop would stop after run-3
  (second overall fail) — so fewer than 5 lines means reset is broken.
- Non-zero exit (last run failed).

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
		t.Fatalf("expected non-zero (last run fails); stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assert.Output(t, resp.Stdout, `---
version: 2
---
run-1
run-2
run-3
run-4
run-5
`)
}
```
