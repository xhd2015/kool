//go:build ignore

package dev

import (
	"os"
	"reflect"
	"testing"
)

func TestParseArgsForwardsSharedFlagsAndNormalizesPrefix(t *testing.T) {
	prefix, prefixSet, args, err := parseArgs([]string{
		"--port", "8080", "--route-prefix", " /demo// ", "--no-air", "--no-open",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !prefixSet || prefix != "/demo" {
		t.Fatalf("prefix = %q, set = %t, want /demo and true", prefix, prefixSet)
	}
	want := []string{"--port", "8080", "--no-use-air", "--no-open"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
}

func TestParseArgsStopsAtDoubleDash(t *testing.T) {
	prefix, prefixSet, args, err := parseArgs([]string{"--route-prefix=demo", "--", "--no-air"})
	if err != nil {
		t.Fatal(err)
	}
	if !prefixSet || prefix != "/demo" {
		t.Fatalf("prefix = %q, set = %t", prefix, prefixSet)
	}
	if want := []string{"--", "--no-air"}; !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
}

func TestSetRoutePrefixEnvRestoresExistingValue(t *testing.T) {
	previous, existed := os.LookupEnv(routePrefixEnv)
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(routePrefixEnv, previous)
		} else {
			_ = os.Unsetenv(routePrefixEnv)
		}
	})
	if err := os.Setenv(routePrefixEnv, "/old"); err != nil {
		t.Fatal(err)
	}
	restore, err := setRoutePrefixEnv("/demo")
	if err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv(routePrefixEnv); got != "/demo" {
		t.Fatalf("env = %q", got)
	}
	restore()
	if got := os.Getenv(routePrefixEnv); got != "/old" {
		t.Fatalf("restored env = %q", got)
	}
}

func TestRoutePrefixArgs(t *testing.T) {
	if got := routePrefixViteArgs(""); got != nil {
		t.Fatalf("root Vite args = %#v, want nil", got)
	}
	if got, want := routePrefixViteArgs("/demo"), []string{"--base", "/demo/"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Vite args = %#v, want %#v", got, want)
	}
	if got := routePrefixBrowserPath("/demo"); got != "/demo/" {
		t.Fatalf("browser path = %q", got)
	}
}
