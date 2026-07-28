# Scenario

**Feature**: empty tabs array is rejected

```
{"version":1,"tabs":[]} -> show emptytabs -> Error
```

## Steps

1. Write emptytabs.json with empty tabs.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	writeConfigFile(t, req.ConfigDir, "emptytabs", `{
  "version": 1,
  "tabs": []
}
`)
	req.SetName = "emptytabs"
	req.Subcommand = "show"
	return nil
}
```
