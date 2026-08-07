# Scenario

**Feature**: home-linked layer order is real HOME seed < primary pack < load (later wins)

```
# fake home shared.txt=H; primary shared.txt=P; load shared.txt=L; --home-linked
HOME=FAKE KOOL_SANDBOX_ROOT=PARENT ./primary --load-devbox ABS -- \
  sh -c 'cat shared.txt; printf HOME=%s\nSANDBOX=%s\n "$HOME" "$SANDBOX_ROOT"'
  -> exit 0; shared.txt == L; HOME == SANDBOX_ROOT (session)
```

## Steps

1. Fake real home seeds `shared.txt` content H.
2. Primary `--home-linked` packs `shared.txt` content P.
3. Secondary packs `shared.txt` content L; load via SealedLoadDevbox.
4. Guest cats shared.txt and prints HOME/SANDBOX_ROOT.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	home, err := writeFakeHome(t, req.WorkingDir, "fake-real-home", map[string]string{
		"shared.txt": "content-H-from-real-home\n",
	})
	if err != nil {
		return err
	}
	req.SealedHome = home
	req.HomeLinked = true

	if _, err := writeLocalFile(t, req.WorkingDir, "primary-shared.txt", "content-P-from-primary\n"); err != nil {
		return err
	}
	if _, err := writeLocalFile(t, req.WorkingDir, "load-shared.txt", "content-L-from-load\n"); err != nil {
		return err
	}
	secOut := "load-home-linked.bin"
	secAbs := filepath.Join(req.WorkingDir, secOut)
	req.SecondaryPacks = []SecondaryPack{{
		Output:     secOut,
		ExtraFiles: []string{"load-shared.txt=shared.txt"},
	}}
	req.ExtraFiles = []string{"primary-shared.txt=shared.txt"}
	req.SealedLoadDevbox = []string{secAbs}
	req.SealedArgs = []string{"sh", "-c", `cat shared.txt; printf 'HOME=%s\n' "$HOME"; printf 'SANDBOX=%s\n' "$SANDBOX_ROOT"`}
	return nil
}
```
