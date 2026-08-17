# Scenario

**Feature**: missing dest with Download true calls Install(pin) after writing Prompt

```
# dest missing, hook stands in for downloadgo.Download
missing dest + Download:true + Prompt -> Stderr Prompt -> Install(go1.19.13, installDir) -> hook path
```

## Steps

1. Leave dest missing. Set `Download` true.
2. Set Prompt and a hook return path under a separate temp dir.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Download = true
	req.Prompt = "installing go1.19.13\n"
	req.HookGoroot = filepath.Join(t.TempDir(), "hook-sdk")
	return nil
}
```
