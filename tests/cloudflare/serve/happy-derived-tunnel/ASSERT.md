## Expected

- Exit 0.
- StartSession called with Domain `a.example.com`, LocalURL `http://127.0.0.1:9`,
  TunnelName `kool-lb-a`.
- WaitSignal called; Stop called.
- Stdout contains public URL `https://a.example.com`.

## Exit Code

- 0

```go
import (
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
	if !resp.StartCalled {
		t.Fatal("expected StartSession to be called")
	}
	if resp.StartDomain != "a.example.com" {
		t.Fatalf("StartDomain=%q want a.example.com", resp.StartDomain)
	}
	if resp.StartLocalURL != "http://127.0.0.1:9" {
		t.Fatalf("StartLocalURL=%q want http://127.0.0.1:9", resp.StartLocalURL)
	}
	if resp.StartTunnelName != "kool-lb-a" {
		t.Fatalf("StartTunnelName=%q want kool-lb-a", resp.StartTunnelName)
	}
	if !resp.WaitSignalCalled {
		t.Fatal("expected WaitSignal to be called")
	}
	if !resp.StopCalled {
		t.Fatal("expected Session.Stop to be called after WaitSignal")
	}
	// Public URL required; tunnel name on stdout matches product mockup (soft multi-token).
	assert.Output(t, resp.Stdout, `<contains>
https://a.example.com
kool-lb-a
</contains>
`)
}
```
