## Expected

- Exit 1 after fake osascript fails.

## Exit Code

- 1

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected osascript failure exit, stderr=%s", resp.Stderr)
	}
}
```