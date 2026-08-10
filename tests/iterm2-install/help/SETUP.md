# Scenario

**Feature**: install help documents `--via-open`

```
# user asks for install help
user -> kool iterm2 install --help
  -> usage on stdout, exit 0 (no resolve / no download)
```

## Steps

1. Mark Help; no fake HTTP.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Help = true
	req.UseFakeHTTP = false
	req.DownloadDir = "" // help: no download-dir needed
	return nil
}
```
