# Scenario

**Feature**: structured snapshot formats remain fully buffered (not progressive)

```
sessions snapshot --json|--markdown|--html|-o FILE
  -> full Capture then one render
  -> JSON is a single document; not NDJSON window chunks
```

## Steps

1. Mode=snapshot-cli; two-window fixture; NoColor for any CLI fallback.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeSnapshotCLI
	req.UseTwoWindowFixture = true
	req.NoColor = true
	if req.ITermRunning == nil {
		req.ITermRunning = boolPtr(true)
	}
	return nil
}
```
