## Expected

- Exit 0.
- Stdout documents `build` and flags used for packing (`-o` or `--output`, input/
  file/env, cross-compile).
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
		t.Fatal("expected help on stdout")
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("help stdout must end with newline; got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `<contains>
build
</contains>
`)
	low := strings.ToLower(resp.Stdout)
	// Accept short or long forms for the principal flags.
	hasOut := strings.Contains(low, "-o") || strings.Contains(low, "--output")
	hasIn := strings.Contains(low, "-i") || strings.Contains(low, "--input") || strings.Contains(low, "--file")
	hasEnv := strings.Contains(low, "--env")
	hasOS := strings.Contains(low, "--goos")
	hasArch := strings.Contains(low, "--goarch")
	if !hasOut || !hasIn || !hasEnv || !hasOS || !hasArch {
		t.Fatalf("root help should mention -o/-i|--file/--env/--goos/--goarch; stdout=%q", resp.Stdout)
	}
}
```
