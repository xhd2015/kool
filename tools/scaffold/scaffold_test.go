package scaffold

import (
	"bytes"
	"strings"
	"testing"
)

func TestHandleList(t *testing.T) {
	var buf bytes.Buffer
	if err := HandleWithWriter(&buf, []string{"--list"}); err != nil {
		t.Fatal(err)
	}
	if got, want := buf.String(), "go-cmd-run-lib\n"; got != want {
		t.Fatalf("list output mismatch:\nwant %q\ngot  %q", want, got)
	}
}

func TestHandleGoCmdRunLib(t *testing.T) {
	var buf bytes.Buffer
	if err := HandleWithWriter(&buf, []string{"go-cmd-run-lib"}); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	for _, want := range []string{
		"# cmd/__NAME__/main.go",
		"# run/__NAME__/run.go",
		"# pkgs/__NAME__/__NAME__.go",
		`__NAME__ "__MODULE__/run/__NAME__"`,
		`core "__MODULE__/pkgs/__NAME__"`,
		"github.com/xhd2015/less-gen/flags",
		"func Run(args []string) error",
		"func Run(config Config) error",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("scaffold output missing %q:\n%s", want, got)
		}
	}
}
