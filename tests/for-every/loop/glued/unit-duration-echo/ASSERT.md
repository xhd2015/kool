## Expected Output

```
hello-glued
hello-glued
```

## Expected

- Exit 0.
- Stdout is exactly two lines of `hello-glued` with trailing newlines.
- Demonstrates glued form routing (`for-every-10ms`) works like spaced.

## Exit Code

- 0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `---
version: 2
---
hello-glued
hello-glued
`)
}
```
