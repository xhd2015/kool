# Scenario

**Feature**: root --help lists build and principal pack flags

```
kool sandbox --help
  -> exit 0; stdout mentions build, -o/--output, -i/--input, --file, --env, --goos, --goarch
```

## Steps

1. HelpAtRoot=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.HelpAtRoot = true
	req.HelpBuild = false
	req.Subcommand = ""
	return nil
}
```
