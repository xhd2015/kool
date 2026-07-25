# Scenario

**Feature**: CLI streaming vs --no-stream buffer for sessions snapshot

```
# default CLI streams window blocks as each window is known
sessions snapshot --no-color
  -> ListTabs(1) -> enrich -> write W1 …
  -> ListTabs(2)  # stdout already has W1 when this starts
  -> write W2 … -> footer

# --no-stream buffers like pre-stream
sessions snapshot --no-stream --no-color
  -> full Capture then one render; W1 not on stdout during last ListTabs
```

## Steps

1. Mode=snapshot-cli; two-window fixture; observe stream order probe.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = ModeSnapshotCLI
	req.UseTwoWindowFixture = true
	req.ObserveStreamOrder = true
	req.NoColor = true
	if req.ITermRunning == nil {
		req.ITermRunning = boolPtr(true)
	}
	return nil
}
```
