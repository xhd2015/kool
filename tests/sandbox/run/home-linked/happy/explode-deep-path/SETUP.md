# Scenario

**Feature**: packing under a seeded directory explodes intermediate symlinks; sibling re-links remain

```
# real home: .config/other/x = O; pack: .config/mytool/cfg = C
kool sandbox build -o sandbox.bin --home-linked --file cfg=.config/mytool/cfg
HOME=FAKE KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- sh -c 'cat .config/mytool/cfg; echo ---; cat .config/other/x'
  -> exit 0; packed C and sibling O both readable (explode re-links .config children)
```

## Steps

1. Fake real home nested path `.config/other/x` with content O.
2. Pack `.config/mytool/cfg` with content C (forces explode of top-level `.config` symlink).
3. Guest cats both packed path and sibling seed path.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	home, err := writeFakeHome(t, req.WorkingDir, "fake-real-home", map[string]string{
		".config/other/x": "content-O-sibling\n",
	})
	if err != nil {
		return err
	}
	req.SealedHome = home

	if _, err := writeLocalFile(t, req.WorkingDir, "cfg", "content-C-packed\n"); err != nil {
		return err
	}
	req.ExtraFiles = []string{"cfg=.config/mytool/cfg"}
	req.SealedArgs = []string{"sh", "-c", `cat .config/mytool/cfg; printf '%s\n' '---'; cat .config/other/x`}
	return nil
}
```
