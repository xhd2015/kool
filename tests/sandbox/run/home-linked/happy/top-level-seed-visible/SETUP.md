# Scenario

**Feature**: top-level real-home entry is visible in the sandbox via seed symlink

```
# fake real home has seed.txt; pack is unrelated
kool sandbox build -o sandbox.bin --home-linked --file packed.txt=packed.txt
HOME=FAKE KOOL_SANDBOX_ROOT=PARENT ./sandbox.bin -- sh -c 'cat seed.txt'
  -> exit 0; stdout == real-home seed content (via top-level symlink)
```

## Steps

1. Fake real home: `seed.txt` with unique content.
2. Pack only `packed.txt` (unrelated path).
3. Guest cats `seed.txt` from sandbox cwd (seeded from real home).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	const seedBody = "seed-from-real-home-unique\n"
	home, err := writeFakeHome(t, req.WorkingDir, "fake-real-home", map[string]string{
		"seed.txt": seedBody,
	})
	if err != nil {
		return err
	}
	req.SealedHome = home

	if _, err := writeLocalFile(t, req.WorkingDir, "packed.txt", "packed-unrelated\n"); err != nil {
		return err
	}
	req.ExtraFiles = []string{"packed.txt=packed.txt"}
	req.SealedArgs = []string{"sh", "-c", "cat seed.txt"}
	return nil
}
```
