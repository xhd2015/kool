# Scenario

**Feature**: list with empty config dir succeeds with zero sets

```
empty KOOL_ITERM2_TAB_SET_DIR -> list -> exit 0, empty or "0 sets"
```

## Steps

1. ConfigDir exists but has no JSON files (root Setup empty dir).

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	// Ensure config dir exists and is empty of *.json.
	entries, err := os.ReadDir(req.ConfigDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		_ = os.RemoveAll(filepath.Join(req.ConfigDir, e.Name()))
	}
	return nil
}
```

