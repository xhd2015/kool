## Expected

- Exit 0.
- Output includes label `Build All`, dependsOn targets `Compile` and `Serve`.
- Prefer type composite (or empty type with deps).

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("show composite exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	if !strings.Contains(out, "Build All") {
		t.Fatalf("show missing Build All; out:\n%s", out)
	}
	if !strings.Contains(out, "Compile") || !strings.Contains(out, "Serve") {
		t.Fatalf("show composite should list dependsOn Compile and Serve; out:\n%s", out)
	}
}
```
