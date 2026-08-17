## Expected

- `IncrementTag` returns no error.
- `NextTag` is `v0.2.1`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("unexpected harness error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("IncrementTag(%q) error = %v, want nil", req.Tag, resp.Err)
	}
	if resp.NextTag != "v0.2.1" {
		t.Fatalf("IncrementTag(%q) = %q, want v0.2.1", req.Tag, resp.NextTag)
	}
}
```