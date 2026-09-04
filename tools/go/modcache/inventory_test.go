package modcache

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestInventoryAndReport(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "example.com", "foo@v1.0.0", "x.go"), "old\n")
	writeFile(t, filepath.Join(root, "example.com", "foo@v1.2.0", "x.go"), "new\n")
	writeFile(t, filepath.Join(root, "cache", "download", "example.com", "foo", "@v", "v1.0.0.zip"), "ZIPOLD")
	writeFile(t, filepath.Join(root, "cache", "download", "example.com", "foo", "@v", "v1.2.0.zip"), "ZIPNEW")
	writeFile(t, filepath.Join(root, "golang.org", "toolchain@v0.0.1-go1.21.0.darwin-arm64", "bin"), "t1")
	writeFile(t, filepath.Join(root, "golang.org", "toolchain@v0.0.1-go1.22.0.darwin-arm64", "bin"), "t2")
	writeFile(t, filepath.Join(root, "github.com", "!azure", "ok@v1.0.0", "a.go"), "az\n")

	inv, err := inventoryCache(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if inv.Modules["example.com/foo"] == nil {
		t.Fatalf("missing example.com/foo: %#v", inv.Modules)
	}
	if inv.Modules["github.com/Azure/ok"] == nil {
		t.Fatalf("unescape failed, modules=%v", keys(inv.Modules))
	}
	if inv.Modules["golang.org/toolchain"] == nil {
		t.Fatal("missing toolchain")
	}

	rep := buildReport(inv, nil, false)
	if rep.LegacyVersions != 1 {
		t.Fatalf("legacy versions=%d want 1 (toolchain excluded)", rep.LegacyVersions)
	}
	if rep.SaveBytes != rep.LegacyBytes || rep.SaveBytes <= 0 {
		t.Fatalf("saveBytes=%d legacyBytes=%d want equal and >0", rep.SaveBytes, rep.LegacyBytes)
	}
	if rep.ToolchainVers != 2 {
		t.Fatalf("toolchain versions=%d want 2", rep.ToolchainVers)
	}
	var foundLegacy bool
	for _, m := range rep.Modules {
		if m.Path == "example.com/foo" {
			if m.Newest != "v1.2.0" {
				t.Fatalf("newest=%s", m.Newest)
			}
			for _, v := range m.Versions {
				if v.Version == "v1.0.0" && v.Legacy {
					foundLegacy = true
				}
			}
		}
		if m.Path == "golang.org/toolchain" {
			for _, v := range m.Versions {
				if v.Legacy {
					t.Fatalf("toolchain version %s marked legacy without --include-toolchain", v.Version)
				}
			}
		}
	}
	if !foundLegacy {
		t.Fatal("expected example.com/foo@v1.0.0 to be legacy")
	}

	repInc := buildReport(inv, nil, true)
	if repInc.LegacyVersions < 2 {
		t.Fatalf("with toolchain, legacy=%d want >=2", repInc.LegacyVersions)
	}
}

func keys(m map[string]*cachedModule) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
