## Expected

- Non-zero exit; stderr mentions usage / directory.

## Exit Code

- 1

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
		t.Fatalf("expected non-zero exit, got 0\nstderr: %s", resp.Stderr)
	}
	assert.Output(t, resp.Stderr+resp.Stdout, `
<contains>
iterm2
</contains>`)
}
```