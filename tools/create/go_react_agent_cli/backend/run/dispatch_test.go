//go:build ignore

package run

import (
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout runs fn while capturing everything written to stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()
	fn()
	w.Close()
	os.Stdout = old
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestRunNoArgsPrintsRootHelp(t *testing.T) {
	out := captureStdout(t, func() {
		if err := Run(nil); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
	if !strings.Contains(out, "Usage: __PROJECT_NAME__ <command>") {
		t.Fatalf("expected root help, got:\n%s", out)
	}
	if !strings.Contains(out, "server") || !strings.Contains(out, "skill") {
		t.Fatalf("expected command list, got:\n%s", out)
	}
}

func TestRunHelpFlagsPrintRootHelp(t *testing.T) {
	for _, arg := range []string{"-h", "--help", "help"} {
		out := captureStdout(t, func() {
			if err := Run([]string{arg}); err != nil {
				t.Fatalf("%s: expected no error, got %v", arg, err)
			}
		})
		if !strings.Contains(out, "Usage: __PROJECT_NAME__ <command>") {
			t.Fatalf("%s: expected root help, got:\n%s", arg, out)
		}
	}
}

func TestRunUnknownCommand(t *testing.T) {
	err := Run([]string{"bogus"})
	if err == nil || !strings.Contains(err.Error(), "unknown command: bogus") {
		t.Fatalf("expected unknown command error, got %v", err)
	}
	if !strings.Contains(err.Error(), "__PROJECT_NAME__ --help") {
		t.Fatalf("expected help hint, got %v", err)
	}
}

func TestServerHelpMentionsDefaultPort(t *testing.T) {
	out := captureStdout(t, func() {
		if err := handleServer([]string{"--help"}); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
	if !strings.Contains(out, "8080") {
		t.Fatalf("expected server help to mention default port 8080, got:\n%s", out)
	}
	if !strings.Contains(out, "--route-prefix") {
		t.Fatalf("expected server flags in help, got:\n%s", out)
	}
}

func TestServerExternalViteRequiresDev(t *testing.T) {
	err := handleServer([]string{"--external-vite"})
	if err == nil || !strings.Contains(err.Error(), "--external-vite requires --dev") {
		t.Fatalf("expected --external-vite error, got %v", err)
	}
}
