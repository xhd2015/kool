## Expected

- Exit 1; usage or unknown argument error.

## Exit Code

- 1

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected failure for extra args, stderr=%s", resp.Stderr)
	}
}
```