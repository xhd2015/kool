## Expected

- Non-zero exit.
- StartSession was attempted.
- Stderr surfaces the start error (contains `simulated start failure` or
  generic start/tunnel wording plus non-empty message).
- WaitSignal and Stop should not complete a successful lifecycle (Stop not required).

## Errors

- StartSession failure.

## Exit Code

- non-zero

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero when StartSession fails; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if !resp.StartCalled {
		t.Fatal("expected StartSession to be attempted")
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatal("expected error on stderr")
	}
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "simulated start failure") &&
		!strings.Contains(low, "start") &&
		!strings.Contains(low, "tunnel") &&
		!strings.Contains(low, "fail") {
		t.Fatalf("stderr should surface start error; got %q", resp.Stderr)
	}
	if resp.WaitSignalCalled {
		t.Fatal("WaitSignal must not run after StartSession failure")
	}
	if resp.StopCalled {
		t.Fatal("Stop must not run when StartSession failed")
	}
}
```
