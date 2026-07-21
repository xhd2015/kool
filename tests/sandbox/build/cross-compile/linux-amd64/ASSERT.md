## Expected

- Exit 0.
- Output binary exists with size > 0.
- Stdout mentions `linux` and `amd64` (target triple).
- If `file` is on PATH, its report on the binary should indicate ELF (soft).

## Exit Code

- 0

```go
import (
	"os/exec"
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
	if !resp.OutputExists || resp.OutputSize <= 0 {
		t.Fatalf("expected sealed binary; exists=%v size=%d path=%q", resp.OutputExists, resp.OutputSize, resp.OutputPath)
	}
	assert.Output(t, resp.Stdout, `<contains>
linux
amd64
</contains>
`)
	// Optional host `file` probe (available on macOS/Linux with file(1)).
	if path, lookErr := exec.LookPath("file"); lookErr == nil {
		out, runErr := exec.Command(path, resp.OutputPath).CombinedOutput()
		if runErr != nil {
			t.Logf("file command failed (ignored): %v out=%q", runErr, out)
			return
		}
		low := strings.ToLower(string(out))
		if !strings.Contains(low, "elf") {
			t.Fatalf("expected ELF binary for linux/amd64; file=%q", out)
		}
	}
}
```
