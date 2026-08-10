# Scenario

**Feature**: `--spaces` after multi-app merge; kept windows retain correct `app`

```
Caller
  -> sessions save --spaces 0 --file app-spaces.json
  -> multi-app spaces fixture: system@space0 + home@space2
  <- keeps space-0 window with system app; skip warn; filter.spaces [0]
```

## Steps

1. UseMultiAppSpacesFixture; Spaces=0; FilePath=app-spaces.json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.UseMultiAppSpacesFixture = true
	req.Spaces = "0"
	req.FilePath = "app-spaces.json"
	return nil
}
```
