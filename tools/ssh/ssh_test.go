package ssh

import (
	"bytes"
	"strings"
	"testing"
)

func TestBuildSSHForwardArgv(t *testing.T) {
	argv := BuildSSHForwardArgv(18082, "127.0.0.1", 8082, "test.devbox")
	want := []string{
		"-N",
		"-o", "ExitOnForwardFailure=yes",
		"-o", "ServerAliveInterval=30",
		"-L", "18082:127.0.0.1:8082",
		"test.devbox",
	}
	if len(argv) != len(want) {
		t.Fatalf("len=%d want %d: %#v", len(argv), len(want), argv)
	}
	for i := range want {
		if argv[i] != want[i] {
			t.Fatalf("argv[%d]=%q want %q\nfull=%#v", i, argv[i], want[i], argv)
		}
	}
}

func TestForwardHelp(t *testing.T) {
	var out bytes.Buffer
	err := HandleWith([]string{"forward", "--help"}, HandleOpts{Stdout: &out, RunSSH: func([]string) error {
		t.Fatal("should not run ssh")
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.HasSuffix(s, "\n") {
		t.Fatalf("help must end with newline")
	}
	for _, sub := range []string{"--local", "--to-remote-internal", "--host", "forward"} {
		if !strings.Contains(s, sub) {
			t.Fatalf("help missing %q\n%s", sub, s)
		}
	}
}

func TestForwardPrintsSSHCommandAndRuns(t *testing.T) {
	var out bytes.Buffer
	var got []string
	err := HandleWith([]string{
		"forward",
		"--local", "18082",
		"--to-remote-internal", "127.0.0.1:8082",
		"--host", "test.devbox",
	}, HandleOpts{
		Stdout: &out,
		RunSSH: func(argv []string) error {
			got = append([]string(nil), argv...)
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("RunSSH not called")
	}
	wantLine := "ssh -N -o ExitOnForwardFailure=yes -o ServerAliveInterval=30 -L 18082:127.0.0.1:8082 test.devbox"
	if !strings.Contains(out.String(), wantLine) {
		t.Fatalf("stdout missing ssh command %q\n%s", wantLine, out.String())
	}
	if !strings.Contains(out.String(), "http://127.0.0.1:18082") {
		t.Fatalf("missing Local URL\n%s", out.String())
	}
}

func TestForwardMissingFlags(t *testing.T) {
	cases := []struct {
		name string
		args []string
		sub  string
	}{
		{"no local", []string{"forward", "--to-remote-internal", "127.0.0.1:8082", "--host", "test.devbox"}, "--local"},
		{"no to", []string{"forward", "--local", "18082", "--host", "test.devbox"}, "--to-remote-internal"},
		{"no host", []string{"forward", "--local", "18082", "--to-remote-internal", "127.0.0.1:8082"}, "--host"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := HandleWith(tc.args, HandleOpts{RunSSH: func([]string) error {
				t.Fatal("should not run")
				return nil
			}})
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.sub) {
				t.Fatalf("err %q should mention %q", err, tc.sub)
			}
		})
	}
}
