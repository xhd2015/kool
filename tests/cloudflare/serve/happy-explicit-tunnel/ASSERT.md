## Expected

- Exit 0.
- StartSession TunnelName is exactly `my-tun` (not derived `kool-lb-a`).
- WaitSignal and Stop called.

## Exit Code

- 0

```go
import "testing"

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
	if resp.StartTunnelName != "my-tun" {
		t.Fatalf("StartTunnelName=%q want my-tun", resp.StartTunnelName)
	}
	if resp.StartDomain != "a.example.com" {
		t.Fatalf("StartDomain=%q want a.example.com", resp.StartDomain)
	}
	if resp.StartLocalURL != "http://127.0.0.1:9" {
		t.Fatalf("StartLocalURL=%q want http://127.0.0.1:9", resp.StartLocalURL)
	}
	if !resp.WaitSignalCalled {
		t.Fatal("expected WaitSignal to be called")
	}
	if !resp.StopCalled {
		t.Fatal("expected Session.Stop to be called")
	}
}
```
