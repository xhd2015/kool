//go:build ignore

package dev

import (
	"os"
	"reflect"
	"testing"

	"__MODULE_NAME__/server"
)

func TestParseArgsForwardsSharedFlagsAndNormalizesPrefix(t *testing.T) {
	flags, args, err := parseArgs([]string{
		"--port", "8080", "--route-prefix", " /demo// ", "--keep-root-route", "--no-air", "--no-open",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !flags.prefixSet || flags.route.Prefix != "/demo" {
		t.Fatalf("route = %+v, want /demo and prefixSet", flags)
	}
	if !flags.keepRootSet || !flags.route.KeepRoot {
		t.Fatalf("keep-root = %+v, want set and true", flags)
	}
	want := []string{"--port", "8080", "--no-use-air", "--no-open"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
}

func TestParseArgsStopsAtDoubleDash(t *testing.T) {
	flags, args, err := parseArgs([]string{"--route-prefix=demo", "--", "--no-air"})
	if err != nil {
		t.Fatal(err)
	}
	if !flags.prefixSet || flags.route.Prefix != "/demo" {
		t.Fatalf("route = %+v", flags)
	}
	if want := []string{"--", "--no-air"}; !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
}

func TestParseArgsReadsKeepRootRouteSpellings(t *testing.T) {
	for _, args := range [][]string{
		{"--keep-root-route"},
		{"--keep-root-route=true"},
		{"--keep-root-route=1"},
	} {
		flags, rest, err := parseArgs(args)
		if err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if !flags.keepRootSet || !flags.route.KeepRoot {
			t.Fatalf("%v: route = %+v", args, flags)
		}
		if len(rest) != 0 {
			t.Fatalf("%v: rest = %#v", args, rest)
		}
	}

	flags, _, err := parseArgs([]string{"--keep-root-route=false"})
	if err != nil {
		t.Fatal(err)
	}
	if !flags.keepRootSet || flags.route.KeepRoot {
		t.Fatalf("false spelling: %+v", flags)
	}
}

func TestParseArgsRejectsInvalidPrefix(t *testing.T) {
	if _, _, err := parseArgs([]string{"--route-prefix", "a b"}); err == nil {
		t.Fatal("invalid prefix should fail")
	}
	if _, _, err := parseArgs([]string{"--route-prefix"}); err == nil {
		t.Fatal("missing value should fail")
	}
}

func TestSetRouteEnvRestoresExistingValues(t *testing.T) {
	previousPrefix, hadPrefix := os.LookupEnv(routePrefixEnv)
	previousKeepRoot, hadKeepRoot := os.LookupEnv(keepRootEnv)
	t.Cleanup(func() {
		restoreEnvVar(routePrefixEnv, previousPrefix, hadPrefix)
		restoreEnvVar(keepRootEnv, previousKeepRoot, hadKeepRoot)
	})
	if err := os.Setenv(routePrefixEnv, "/old"); err != nil {
		t.Fatal(err)
	}
	if err := os.Unsetenv(keepRootEnv); err != nil {
		t.Fatal(err)
	}

	restore, err := setRouteEnv(server.RouteOptions{Prefix: "/demo", KeepRoot: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv(routePrefixEnv); got != "/demo" {
		t.Fatalf("prefix env = %q", got)
	}
	if got := os.Getenv(keepRootEnv); got != "true" {
		t.Fatalf("keep-root env = %q", got)
	}

	restore()
	if got := os.Getenv(routePrefixEnv); got != "/old" {
		t.Fatalf("restored prefix env = %q", got)
	}
	if _, ok := os.LookupEnv(keepRootEnv); ok {
		t.Fatalf("keep-root env should be unset again, got %q", os.Getenv(keepRootEnv))
	}
}

func restoreEnvVar(key, value string, existed bool) {
	if existed {
		_ = os.Setenv(key, value)
		return
	}
	_ = os.Unsetenv(key)
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
