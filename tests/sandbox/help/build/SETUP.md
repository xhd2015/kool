# Scenario

**Feature**: build --help documents pack flags including --home-linked and --runtime-load-devbox

```
kool sandbox build --help
  -> exit 0; stdout documents -o/-i/--file/--env/--goos/--goarch/--home-linked/--runtime-load-devbox
```


## Steps

1. HelpBuild=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.HelpAtRoot = false
	req.HelpBuild = true
	req.Subcommand = ""
	return nil
}
```
