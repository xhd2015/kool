## Expected Output

```
  sandbox   ...
  files     1
  env       1
  size      ...
```

(soft match; gray/meta optional; ends with newline)

## Expected

- Exit 0.
- Output binary exists with size > 0.
- Stdout mentions files and env counts (at least one file and one env).
- Stdout ends with `\n`.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d want 0; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !resp.OutputExists {
		t.Fatalf("expected output binary at %q", resp.OutputPath)
	}
	if resp.OutputSize <= 0 {
		t.Fatalf("expected output size > 0; got %d path=%q", resp.OutputSize, resp.OutputPath)
	}
	if resp.Stdout == "" {
		t.Fatal("expected build summary on stdout")
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("build stdout must end with newline; got %q", resp.Stdout)
	}
	// Soft multi-token: product may use gray labels; require files/env keywords + counts.
	assert.Output(t, resp.Stdout, `<contains>
files
env
</contains>
`)
	low := strings.ToLower(resp.Stdout)
	if !strings.Contains(low, "1") {
		t.Fatalf("stdout should include file/env count; got %q", resp.Stdout)
	}
}
```
