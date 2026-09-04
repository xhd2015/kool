package modcache

import "testing"

func TestFormatBytes(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{0, "0B"},
		{512, "512B"},
		{1023, "1023B"},
		{1024, "1K"},
		{1536, "1.5K"},
		{1024 * 1024, "1M"},
		{398 * 1024 * 1024, "398M"},
		{7*1024*1024*1024 + 300*1024*1024, "7.3G"},
	}
	for _, tc := range cases {
		if got := formatBytes(tc.n); got != tc.want {
			t.Errorf("formatBytes(%d)=%q want %q", tc.n, got, tc.want)
		}
	}
}

func TestParseDownloadFile(t *testing.T) {
	cases := []struct {
		name    string
		version string
		kind    string
		ok      bool
	}{
		{"v1.0.0.zip", "v1.0.0", "zip", true},
		{"v1.0.0.ziphash", "v1.0.0", "ziphash", true},
		{"v1.0.0.info", "v1.0.0", "info", true},
		{"v1.0.0.mod", "v1.0.0", "mod", true},
		{"v1.0.0.lock", "v1.0.0", "lock", true},
		{"v1.0.0.partial", "v1.0.0", "partial", true},
		{"list", "", "", false},
	}
	for _, tc := range cases {
		ver, kind, ok := parseDownloadFile(tc.name)
		if ok != tc.ok || ver != tc.version || kind != tc.kind {
			t.Errorf("parseDownloadFile(%q)=(%q,%q,%v) want (%q,%q,%v)",
				tc.name, ver, kind, ok, tc.version, tc.kind, tc.ok)
		}
	}
}

func TestNewestVersion(t *testing.T) {
	got := newestVersion([]string{"v1.0.0", "v1.2.0", "v1.1.0"})
	if got != "v1.2.0" {
		t.Fatalf("newest=%q want v1.2.0", got)
	}
	got = newestVersion([]string{"v0.0.0-20211015210444-4f30a5c0130f", "v0.54.0"})
	if got != "v0.54.0" {
		t.Fatalf("newest pseudo vs tagged=%q want v0.54.0", got)
	}
}

func TestIsToolchain(t *testing.T) {
	if !isToolchain("golang.org/toolchain") {
		t.Fatal("expected toolchain path")
	}
	if isToolchain("golang.org/x/tools") {
		t.Fatal("x/tools is not toolchain")
	}
}

func TestSavePercent(t *testing.T) {
	if got := savePercent(0, 100); got != "" {
		t.Fatalf("zero save: %q", got)
	}
	if got := savePercent(50, 0); got != "" {
		t.Fatalf("zero total: %q", got)
	}
	if got := savePercent(1, 1000); got != " (<1% of total)" {
		t.Fatalf("tiny: %q", got)
	}
	if got := savePercent(27, 100); got != " (27% of total)" {
		t.Fatalf("pct: %q", got)
	}
}

func TestIsUnder(t *testing.T) {
	if !isUnder("/tmp/mod/foo", "/tmp/mod") {
		t.Fatal("child should be under root")
	}
	if !isUnder("/tmp/mod", "/tmp/mod") {
		t.Fatal("root should be under itself")
	}
	if isUnder("/tmp/other", "/tmp/mod") {
		t.Fatal("sibling should not be under root")
	}
}
