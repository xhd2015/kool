## Expected

- Exit 0.
- Output binary exists with size > 0.
- Stdout ends with `\n` and mentions files/env (soft).

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
		t.Fatalf("exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !resp.OutputExists || resp.OutputSize <= 0 {
		t.Fatalf("expected sealed binary; exists=%v size=%d path=%q", resp.OutputExists, resp.OutputSize, resp.OutputPath)
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("build stdout must end with newline; got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `<contains>
files
env
</contains>
`)
}
```
