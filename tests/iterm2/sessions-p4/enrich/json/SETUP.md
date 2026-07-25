# Scenario

**Feature**: --json includes agent.session_id (and tree) on enriched sessions

```
sessions snapshot --json
  + busyGrokResolve on ttys002
  -> single JSON document
  -> busy session has agent.kind=grok, agent.session_id=fixture uuid
  -> agent.tree present with node pids when tree not disabled
```

## Steps

1. JSON=true; default busyGrokResolve; enrich + tree on.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.JSON = true
	req.NoColor = true
	req.NoEnrich = false
	req.NoTree = false
	return nil
}
```
