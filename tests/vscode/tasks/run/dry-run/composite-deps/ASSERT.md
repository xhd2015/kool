## Expected

- Exit 0.
- Plan mentions root `Build All` and dependency steps `Compile` and `Serve`.
- Prefer leaf commands present (`go build`, `bin/app`).

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
		t.Fatalf("dry-run composite exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout
	for _, want := range []string{"Build All", "Compile", "Serve"} {
		if !strings.Contains(out, want) {
			t.Fatalf("composite plan missing %q; out:\n%s", want, out)
		}
	}
}
```
