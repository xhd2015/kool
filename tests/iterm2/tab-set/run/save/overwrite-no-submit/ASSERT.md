## Expected

- Exit 0.
- Written bots.json: tab a has `"no_submit": true`; tab b does not claim true.
- Prefer stdout/stderr mentioning `modified` (diff bucket) for the no_submit change;
  file content is the hard requirement.

## Side Effects

- Overwrites bots.json under ConfigDir.

## Exit Code

- 0

```go
import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d

	if err != nil {
		t.Fatal(err)
	}
	out := combinedOut(resp)
	if resp.ExitCode != 0 {
		t.Fatalf("overwrite no_submit exit=%d out:\n%s", resp.ExitCode, out)
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		t.Fatalf("--tab/--save/--force not accepted; out:\n%s", out)
	}
	if strings.Contains(lower, "unknown key") && strings.Contains(lower, "no_submit") {
		t.Fatalf("no_submit prop not accepted; out:\n%s", out)
	}

	data, readErr := os.ReadFile(configPath(req.ConfigDir, "bots"))
	if readErr != nil {
		t.Fatalf("read bots.json: %v", readErr)
	}
	var file struct {
		Version int `json:"version"`
		Tabs    []struct {
			ID       string `json:"id"`
			Command  string `json:"command"`
			NoSubmit *bool  `json:"no_submit"`
		} `json:"tabs"`
	}
	if jerr := json.Unmarshal(data, &file); jerr != nil {
		t.Fatalf("json: %v\n%s", jerr, data)
	}
	if file.Version != 1 {
		t.Fatalf("version want 1 got %d", file.Version)
	}
	byID := map[string]*bool{}
	for i := range file.Tabs {
		byID[file.Tabs[i].ID] = file.Tabs[i].NoSubmit
	}
	if byID["a"] == nil || !*byID["a"] {
		t.Fatalf("tab a want no_submit=true after overwrite; file=%s", data)
	}
	if byID["b"] != nil && *byID["b"] {
		t.Fatalf("tab b must not be no_submit=true; file=%s", data)
	}
	// Prefer modified bucket once no_submit participates in diff equality.
	if !strings.Contains(lower, "modified") {
		t.Logf("note: diff did not mention modified (file content is authoritative): %s", out)
	}
}
```