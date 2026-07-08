## Expected

- `IncrementTag` returns no error.
- `NextTag` is `v0.2.2`.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected harness error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("IncrementTag(%q) error = %v, want nil", req.Tag, resp.Err)
	}
	if resp.NextTag != "v0.2.2" {
		t.Fatalf("IncrementTag(%q) = %q, want v0.2.2", req.Tag, resp.NextTag)
	}
}
```