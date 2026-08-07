# Scenario

**Feature**: guest HOME equals SANDBOX_ROOT (session materialize root) when home-linked

```
kool sandbox build -o sandbox.bin --home-linked --file packed.txt=packed.txt
HOME=FAKE KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- sh -c 'printf HOME/SANDBOX_ROOT'
  -> exit 0; HOME line == SANDBOX_ROOT line == session under PARENT (not FAKE)
```

## Steps

1. Fake real home may hold noise (seeded but not asserted here).
2. Pack an unrelated file so the pack is non-empty.
3. Guest prints `$HOME` then `$SANDBOX_ROOT` on separate lines.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	home, err := writeFakeHome(t, req.WorkingDir, "fake-real-home", map[string]string{
		"noise.txt": "real-home-noise\n",
	})
	if err != nil {
		return err
	}
	req.SealedHome = home

	if _, err := writeLocalFile(t, req.WorkingDir, "packed.txt", "packed-body\n"); err != nil {
		return err
	}
	req.ExtraFiles = []string{"packed.txt=packed.txt"}
	req.SealedArgs = []string{"sh", "-c", `printf '%s\n' "$HOME"; printf '%s\n' "$SANDBOX_ROOT"`}
	return nil
}
```
