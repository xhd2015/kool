## Expected

- Exit 0.
- Output includes tab ids `a` and `b`, commands `echo a` / `echo b`,
  and preferably window name `local-bots`.

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
		t.Fatalf("exit=%d stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	out := resp.Stdout + resp.Stderr
	for _, want := range []string{"echo a", "echo b"} {
		if !strings.Contains(out, want) {
			t.Fatalf("show missing %q; out:\n%s", want, out)
		}
	}
	// tab ids (as standalone tokens)
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Fatalf("show should include tab ids; out:\n%s", out)
	}
	if !strings.Contains(out, "local-bots") {
		t.Fatalf("show should include window_name local-bots; out:\n%s", out)
	}
}
```
