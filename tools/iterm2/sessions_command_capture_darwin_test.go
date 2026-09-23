//go:build darwin

package iterm2

import (
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func TestForegroundDarwinExactArgv(t *testing.T) {
	if os.Getenv("KOOL_FOREGROUND_ARGV_HELPER") == "1" {
		return
	}
	// A waiting copy of the test executable provides stable known argv,
	// including an empty argument, without interpreting those arguments.
	args := []string{"-test.run=TestForegroundDarwinArgvHelper", "--", "space arg", "", "quote'\"", "line\nbreak"}
	cmd := exec.Command(os.Args[0], args...)
	cmd.Env = append(os.Environ(), "KOOL_FOREGROUND_ARGV_HELPER=1", "FOREGROUND_SECRET=must-not-be-argv")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { stdin.Close(); cmd.Wait() }()
	ready := make([]byte, 1)
	if _, err := stdout.Read(ready); err != nil {
		t.Fatal(err)
	}
	identity, err := readForegroundIdentity(cmd.Process.Pid)
	if err != nil || identity.PID != cmd.Process.Pid || identity.PPID != os.Getpid() || identity.StartSec <= 0 {
		t.Fatalf("identity %+v err=%v", identity, err)
	}
	got, err := readForegroundArgv(cmd.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}
	want := append([]string{os.Args[0]}, args...)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argv got %q want %q", got, want)
	}
	cwd, err := readForegroundCwd(cmd.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}
	wantCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if cwd != wantCwd {
		t.Fatalf("cwd %q want %q", cwd, wantCwd)
	}
}

func TestForegroundDarwinArgvHelper(t *testing.T) {
	if os.Getenv("KOOL_FOREGROUND_ARGV_HELPER") != "1" {
		return
	}
	os.Stdout.Write([]byte("!"))
	os.Stdin.Read(make([]byte, 1))
}
