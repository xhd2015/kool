## Expected

- Build exit 0; sealed run non-zero.
- RunStderr Error: style with **seal/unseal/payload/invalid sealed** wording (not
  merely the flag name `load-devbox`).
- Must **not** treat `--load-devbox` as the guest command.

## Errors

- Absolute path exists but is not a sealed sandbox binary (flag parsed; unseal fails).

## Exit Code

- build: 0
- sealed run: non-zero

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("build exit=%d want 0; stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if !resp.RunExecuted {
		t.Fatal("expected sealed binary run")
	}
	if resp.RunExitCode == 0 {
		t.Fatalf("expected non-zero for not-sealed load target; stdout=%q stderr=%q",
			resp.RunStdout, resp.RunStderr)
	}
	if strings.TrimSpace(resp.RunStderr) == "" {
		t.Fatal("expected Error: on sealed stderr")
	}
	if !strings.Contains(resp.RunStderr, "Error:") {
		t.Fatalf("stderr should use Error: prefix; got %q", resp.RunStderr)
	}
	// Vacuous GREEN guard: runner must parse --load-devbox, not exec it as guest.
	if isLoadDevboxExecAsCommand(resp.RunStderr) {
		t.Fatalf("stderr looks like guest exec of --load-devbox (flag not parsed); got %q", resp.RunStderr)
	}
	low := strings.ToLower(resp.RunStderr)
	// Require seal/unseal/payload/invalid sealed sense — not flag name alone.
	// Note: "seal" matches sealed/unseal/unsealed.
	hasSeal := strings.Contains(low, "seal") // seal, sealed, unseal, unsealed
	hasPayload := strings.Contains(low, "payload") || strings.Contains(low, "packblob") || strings.Contains(low, "pack blob")
	hasInvalid := strings.Contains(low, "invalid") || strings.Contains(low, "corrupt") || strings.Contains(low, "malformed")
	hasNotSealed := strings.Contains(low, "not sealed") || strings.Contains(low, "not a sealed") ||
		strings.Contains(low, "not a sandbox")
	if !hasSeal && !hasPayload && !hasInvalid && !hasNotSealed {
		t.Fatalf("stderr should mention seal/unseal/payload/invalid sealed (not merely load-devbox); got %q", resp.RunStderr)
	}
}
```
