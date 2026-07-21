## Expected

- Exit 0.
- Stdout documents `-o`/`--output`, `-i`/`--input`, `--file`, `--env`, `--goos`,
  `--goarch`.
- Stdout ends with newline.

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
		t.Fatalf("exit=%d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stdout == "" {
		t.Fatal("expected build help on stdout")
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("help stdout must end with newline; got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `<contains>
--file
--env
--goos
--goarch
</contains>
`)
	low := strings.ToLower(resp.Stdout)
	if !strings.Contains(low, "-o") && !strings.Contains(low, "--output") {
		t.Fatalf("build help should document -o/--output; stdout=%q", resp.Stdout)
	}
	if !strings.Contains(low, "-i") && !strings.Contains(low, "--input") {
		t.Fatalf("build help should document -i/--input; stdout=%q", resp.Stdout)
	}
}
```
