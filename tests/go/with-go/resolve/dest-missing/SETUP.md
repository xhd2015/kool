# Scenario

**Feature**: dest `$InstallDir/go1.19.13` is absent

```
# no dest dir; Download decides hook vs error
missing dest -> Download false: error | Download true: Prompt then Install hook
```

## Steps

1. Leave dest missing (do not create `$InstallDir/go1.19.13`).
2. Set a default Prompt leaves may override.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Prompt = "should-not-print\n"
	return nil
}
```
