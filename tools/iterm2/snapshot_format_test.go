package iterm2

import (
	"strings"
	"testing"
)

func TestResolveFormat(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		flags   FormatFlags
		path    string
		want    OutputFormat
		wantErr bool
	}{
		{"default", FormatFlags{}, "", FormatCLI, false},
		{"json flag", FormatFlags{JSON: true}, "", FormatJSON, false},
		{"md flag", FormatFlags{Markdown: true}, "", FormatMarkdown, false},
		{"html flag", FormatFlags{HTML: true}, "", FormatHTML, false},
		{"flag wins over suffix", FormatFlags{JSON: true}, "x.md", FormatJSON, false},
		{"suffix json", FormatFlags{}, "a/b.json", FormatJSON, false},
		{"suffix md", FormatFlags{}, "x.MD", FormatMarkdown, false},
		{"suffix html", FormatFlags{}, "out.html", FormatHTML, false},
		{"unknown suffix", FormatFlags{}, "out.txt", FormatCLI, false},
		{"conflict", FormatFlags{JSON: true, HTML: true}, "", "", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := ResolveFormat(tc.flags, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestRedactCommandLine(t *testing.T) {
	t.Parallel()
	in := "spl serve --app-secret=cGafj-secret --port 1"
	out := redactCommandLine(in)
	if out == in || !strings.Contains(out, "***") {
		t.Fatalf("not redacted: %q", out)
	}
	if strings.Contains(out, "cGafj") {
		t.Fatalf("secret leaked: %q", out)
	}
}
