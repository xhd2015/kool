# Scenario

**Feature**: packed top-level file replaces a same-name seed symlink (no content merge)

```
# fake home marker.txt = A; pack marker.txt = B
kool sandbox build -o sandbox.bin --home-linked --file marker.txt=marker.txt
HOME=FAKE KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- sh -c 'cat marker.txt'
  -> exit 0; stdout == B (packed content; seed link replaced)
```

## Steps

1. Fake real home: `marker.txt` content A.
2. Pack `marker.txt` content B via `--file`.
3. Guest cats `marker.txt` — must see packed B, not seed A.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	home, err := writeFakeHome(t, req.WorkingDir, "fake-real-home", map[string]string{
		"marker.txt": "content-A-from-real-home\n",
	})
	if err != nil {
		return err
	}
	req.SealedHome = home

	if _, err := writeLocalFile(t, req.WorkingDir, "marker.txt", "content-B-from-pack\n"); err != nil {
		return err
	}
	req.ExtraFiles = []string{"marker.txt=marker.txt"}
	req.SealedArgs = []string{"sh", "-c", "cat marker.txt"}
	return nil
}
```
