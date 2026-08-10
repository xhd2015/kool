## Expected

- Build exit 0; secondary path present; sock existed after start (reload needs
  a live session listener).
- Notify ran (NotifyRan); prefer NotifyExitCode 0 once publisher exists.
- Session file `reload-me.txt` exists under session root and contains
  `new-load-content`.
- Must not still be only `old-load-content` without the new marker.

## Exit Code

- build: 0; notify: 0 (when implemented)

```go
import (
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
		t.Fatal("expected sealed primary")
	}
	if len(resp.SecondaryPaths) < 1 {
		t.Fatal("expected secondary load path")
	}
	if !resp.SockExistsAfterStart {
		t.Fatalf("hot-reload requires live events sock; parent=%q events=%q",
			resp.SandboxRootParent, resp.EventsDir)
	}
	if !resp.NotifyRan {
		t.Fatal("expected post-start notify-event")
	}
	if resp.NotifyExitCode != 0 {
		t.Fatalf("notify exit=%d want 0; stderr=%q stdout=%q",
			resp.NotifyExitCode, resp.NotifyStderr, resp.NotifyStdout)
	}
	if resp.SessionRoot == "" {
		t.Fatal("expected SessionRoot for file snapshot")
	}
	if !resp.SessionFileExists {
		t.Fatalf("expected session file reload-me.txt under %q", resp.SessionRoot)
	}
	if !strings.Contains(resp.SessionFileContent, "new-load-content") {
		// RED until runner re-applies load Files on devbox.updated.
		t.Fatalf("session file should contain new-load-content; got %q (session=%q)",
			resp.SessionFileContent, resp.SessionRoot)
	}
	// Ensure we are not stuck on old-only content without new marker.
	if strings.Contains(resp.SessionFileContent, "old-load-content") &&
		!strings.Contains(resp.SessionFileContent, "new-load-content") {
		t.Fatalf("session file still old content only: %q", resp.SessionFileContent)
	}
}
```
