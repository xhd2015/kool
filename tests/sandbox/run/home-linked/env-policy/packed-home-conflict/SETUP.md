# Scenario

**Feature**: packed HOME that is not the sandbox session root fails under home-linked

```
kool sandbox build -o sandbox.bin --home-linked --file p.txt=p.txt --env HOME=/tmp/not-sandbox
HOME=FAKE KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- sh -c 'true'
  -> non-zero sealed exit; stderr mentions HOME and/or home-linked (not silent success)
```

## Steps

1. Pack a file plus `--env HOME=/tmp/not-sandbox` (conflicts with home-linked HOME).
2. Sealed run with a trivial guest command — must fail before/when applying packed HOME.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	home, err := writeFakeHome(t, req.WorkingDir, "fake-real-home", nil)
	if err != nil {
		return err
	}
	req.SealedHome = home

	if _, err := writeLocalFile(t, req.WorkingDir, "p.txt", "pack\n"); err != nil {
		return err
	}
	req.ExtraFiles = []string{"p.txt=p.txt"}
	req.ExtraEnv = []string{"HOME=/tmp/not-sandbox"}
	req.SealedArgs = []string{"sh", "-c", "true"}
	return nil
}
```
