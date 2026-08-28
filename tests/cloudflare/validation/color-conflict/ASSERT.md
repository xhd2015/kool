## Expected

- Non-zero exit.
- Stderr mentions `--color and --no-color cannot be specified together`.
- StartSession not called.

## Exit Code

- non-zero

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.StartCalled {
		t.Fatal("StartSession must not run on color conflict")
	}
	if !strings.Contains(resp.Stderr, "--color and --no-color cannot be specified together") {
		t.Fatalf("stderr missing conflict message: %q", resp.Stderr)
	}
}
```
