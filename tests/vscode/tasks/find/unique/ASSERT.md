## Expected

- Exit 0.
- Output includes `Compile`; does not list unrelated labels as the only match noise (Serve may be absent).

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
		t.Fatalf("find unique exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "Compile") {
		t.Fatalf("find should include Compile; out:\n%s", resp.Stdout)
	}
	// Unique match: Serve should not appear as a match row
	if strings.Contains(resp.Stdout, "Serve") {
		t.Fatalf("unique find for compile should not list Serve; out:\n%s", resp.Stdout)
	}
}
```
