//go:build ignore

// Package dev configures the shared development supervisor for this project.
package dev

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"__MODULE_NAME__/server"
	devserver "github.com/xhd2015/dot-pkgs/go-pkgs/dev/server"
)

const (
	routePrefixEnv = "KOOL_GO_REACT_ROUTE_PREFIX"
	keepRootEnv    = "KOOL_GO_REACT_KEEP_ROOT_ROUTE"
	routeFlagHelp  = "  --route-prefix PREFIX  Mount UI and API below PREFIX, e.g. demo\n" +
		"  --keep-root-route      Also serve the unprefixed root route (direct domains)\n"
)

// Run parses project-specific development flags and delegates process lifecycle
// management to dot-pkgs' Go/Air and Vite supervisor.
func Run(args []string) error {
	flags, args, err := parseArgs(args)
	if err != nil {
		return err
	}
	route := flags.route
	if !flags.prefixSet {
		route.Prefix = os.Getenv(routePrefixEnv)
	}
	if !flags.keepRootSet {
		if value, err := strconv.ParseBool(strings.TrimSpace(os.Getenv(keepRootEnv))); err == nil {
			route.KeepRoot = value
		}
	}
	route = server.NormalizeRoute(route)
	restoreEnv, err := setRouteEnv(route)
	if err != nil {
		return err
	}
	defer restoreEnv()

	root, err := devserver.FindRoot("__PROJECT_NAME__-react/package.json")
	if err != nil {
		return err
	}

	err = devserver.Main(args, devserver.Config{
		Name:         "__PROJECT_NAME__",
		Root:         root,
		BuildPackage: "./script/dev",
		BackendPort:  8080,
		FrontendPort: server.DefaultVitePort,
		BrowserPath:  routePrefixBrowserPath(route.Prefix),
		WatchDirs:    []string{"internal/dev", "server", "script/dev"},
		Frontend: &devserver.Vite{
			Dir:        "__PROJECT_NAME__-react",
			ConfigFile: "vite.config.ts",
			Install:    []string{"bun", "install"},
			Command: append([]string{"bun", "run", "dev", "--"},
				routePrefixViteArgs(route.Prefix)...),
		},
		BackendHandler: func(_ context.Context, backend devserver.Backend) (http.Handler, error) {
			return server.DevHandler(backend.FrontendURL, route)
		},
	})
	if err == nil && hasHelp(args) {
		// dot-pkgs owns the shared flags; append this template's one extra flag.
		fmt.Fprint(os.Stdout, routeFlagHelp)
	}
	return err
}

// routeFlags is what this entrypoint takes out of argv before the shared
// supervisor sees the rest.
type routeFlags struct {
	route       server.RouteOptions
	prefixSet   bool
	keepRootSet bool
}

func parseArgs(args []string) (routeFlags, []string, error) {
	var flags routeFlags
	var remaining []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			flags, err := flags.normalize()
			if err != nil {
				return flags, nil, err
			}
			return flags, append(remaining, args[i:]...), nil
		}
		switch {
		case arg == "--route-prefix":
			if i+1 >= len(args) {
				return flags, nil, fmt.Errorf("--route-prefix requires a value")
			}
			i++
			flags.route.Prefix = args[i]
			flags.prefixSet = true
		case strings.HasPrefix(arg, "--route-prefix="):
			flags.route.Prefix = strings.TrimPrefix(arg, "--route-prefix=")
			flags.prefixSet = true
		case arg == "--keep-root-route":
			flags.route.KeepRoot = true
			flags.keepRootSet = true
		case strings.HasPrefix(arg, "--keep-root-route="):
			value := strings.TrimPrefix(arg, "--keep-root-route=")
			flags.route.KeepRoot = value == "" || value == "true" || value == "1"
			flags.keepRootSet = true
		case arg == "--no-air":
			// Preserve the original go-react template flag while dot-pkgs uses
			// the more explicit --no-use-air spelling.
			remaining = append(remaining, "--no-use-air")
		default:
			remaining = append(remaining, arg)
		}
	}
	flags, err := flags.normalize()
	if err != nil {
		return flags, nil, err
	}
	return flags, remaining, nil
}

// normalize canonicalizes the parsed route and rejects an unusable prefix, so
// both the `--` passthrough and the normal path agree on what was requested.
func (f routeFlags) normalize() (routeFlags, error) {
	f.route = server.NormalizeRoute(f.route)
	if err := server.ValidateRoutePrefix(f.route.Prefix); err != nil {
		return f, err
	}
	return f, nil
}

// setRouteEnv publishes the effective route to child processes: the supervisor
// restarts this entrypoint without the original arguments, so the flags have to
// survive as environment.
func setRouteEnv(route server.RouteOptions) (func(), error) {
	restorePrefix, err := setEnv(routePrefixEnv, route.Prefix)
	if err != nil {
		return nil, err
	}
	keepRoot := ""
	if route.KeepRoot {
		keepRoot = "true"
	}
	restoreKeepRoot, err := setEnv(keepRootEnv, keepRoot)
	if err != nil {
		restorePrefix()
		return nil, err
	}
	return func() {
		restoreKeepRoot()
		restorePrefix()
	}, nil
}

func setEnv(key, value string) (func(), error) {
	previous, existed := os.LookupEnv(key)
	var err error
	if value == "" {
		err = os.Unsetenv(key)
	} else {
		err = os.Setenv(key, value)
	}
	if err != nil {
		return nil, err
	}
	return func() {
		if existed {
			_ = os.Setenv(key, previous)
		} else {
			_ = os.Unsetenv(key)
		}
	}, nil
}

func hasHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--" {
			return false
		}
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

func routePrefixViteArgs(routePrefix string) []string {
	if routePrefix == "" {
		return nil
	}
	return []string{"--base", routePrefix + "/"}
}

func routePrefixBrowserPath(routePrefix string) string {
	if routePrefix == "" {
		return "/"
	}
	return routePrefix + "/"
}
