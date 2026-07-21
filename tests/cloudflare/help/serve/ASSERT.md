## Expected

- Exit 0.
- Stdout documents serve flags `--domain`, `--url`, `--tunnel`.
- Stdout ends with newline.

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
		t.Fatal("expected serve help on stdout")
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("help stdout must end with newline; got %q", resp.Stdout)
	}
	if resp.StartCalled {
		t.Fatal("StartSession must not run for serve help")
	}
	assert.Output(t, resp.Stdout, `<contains>
--domain
--url
--tunnel
</contains>
`)
}
```
