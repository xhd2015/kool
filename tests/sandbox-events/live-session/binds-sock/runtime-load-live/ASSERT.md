## Expected

- Primary build exit 0; secondary built; sealed binary exists.
- While guest was live (captured after start): `SockExistsAfterStart` true.
- `SockPath` is under `SandboxRootParent/events/` and ends with `.sock`.
- Session id basename of sock matches a session directory under parent when
  materialize still present before exit (optional soft: SessionRoot non-empty
  before stop is ideal; after exit session may be removed).
- Prefer sock mode 0600 and events dir 0700 when modes recorded (soft fail if
  zero/unknown on some FS; hard fail only on wrong non-zero mode).

## Exit Code

- build: 0

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("build exit=%d want 0; stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if !resp.OutputExists {
		t.Fatal("expected sealed primary binary")
	}
	if len(resp.SecondaryPaths) != 1 {
		t.Fatalf("expected 1 secondary; got %v", resp.SecondaryPaths)
	}
	if !resp.SockExistsAfterStart {
		// RED until runner binds events/<id>.sock during live guest.
		t.Fatalf("expected events sock while guest ran; eventsDir=%q parent=%q guestReady=%v sessionRoot=%q",
			resp.EventsDir, resp.SandboxRootParent, resp.GuestReady, resp.SessionRoot)
	}
	if resp.SockPath == "" {
		t.Fatal("SockPath empty")
	}
	wantPrefix := filepath.Join(resp.SandboxRootParent, "events") + string(filepath.Separator)
	if !strings.HasPrefix(resp.SockPath, wantPrefix) {
		t.Fatalf("sock path %q not under %q", resp.SockPath, wantPrefix)
	}
	if !strings.HasSuffix(resp.SockPath, ".sock") {
		t.Fatalf("sock path should end with .sock: %q", resp.SockPath)
	}
	if resp.SockMode != 0 && resp.SockMode != 0o600 {
		t.Fatalf("sock mode=%o want 0600", resp.SockMode)
	}
	if resp.EventsDirMode != 0 && resp.EventsDirMode != 0o700 {
		t.Fatalf("events dir mode=%o want 0700", resp.EventsDirMode)
	}
}
```
