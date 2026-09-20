//go:build ignore

// Package dev configures the shared development supervisor for this project.
package dev

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"__MODULE_NAME__/server"
	devserver "github.com/xhd2015/dot-pkgs/go-pkgs/dev/server"
)

const (
	routePrefixEnv  = "KOOL_GO_REACT_ROUTE_PREFIX"
	routePrefixHelp = "  --route-prefix PREFIX  Mount UI and API below PREFIX, e.g. demo\n"
)

// Run parses project-specific development flags and delegates process lifecycle
// management to dot-pkgs' Go/Air and Vite supervisor.
func Run(args []string) error {
	routePrefix, routePrefixSet, args, err := parseArgs(args)
	if err != nil {
		return err
	}
	if !routePrefixSet {
		routePrefix = server.NormalizeRoutePrefix(os.Getenv(routePrefixEnv))
	}
	restoreEnv, err := setRoutePrefixEnv(routePrefix)
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
		BrowserPath:  routePrefixBrowserPath(routePrefix),
		WatchDirs:    []string{"internal/dev", "server", "script/dev"},
		Frontend: &devserver.Vite{
			Dir:        "__PROJECT_NAME__-react",
			ConfigFile: "vite.config.ts",
			Install:    []string{"bun", "install"},
			Command: append([]string{"bun", "run", "dev", "--"},
				routePrefixViteArgs(routePrefix)...),
		},
		BackendHandler: func(_ context.Context, backend devserver.Backend) (http.Handler, error) {
			return server.DevHandler(backend.FrontendURL, routePrefix)
		},
	})
	if err == nil && hasHelp(args) {
		// dot-pkgs owns the shared flags; append this template's one extra flag.
		fmt.Fprint(os.Stdout, routePrefixHelp)
	}
	return err
}

func parseArgs(args []string) (routePrefix string, routePrefixSet bool, remaining []string, err error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			return routePrefix, routePrefixSet, append(remaining, args[i:]...), nil
		}
		switch {
		case arg == "--route-prefix":
			if i+1 >= len(args) {
				return "", false, nil, fmt.Errorf("--route-prefix requires a value")
			}
			i++
			routePrefix = server.NormalizeRoutePrefix(args[i])
			routePrefixSet = true
		case strings.HasPrefix(arg, "--route-prefix="):
			routePrefix = server.NormalizeRoutePrefix(strings.TrimPrefix(arg, "--route-prefix="))
			routePrefixSet = true
		case arg == "--no-air":
			// Preserve the original go-react template flag while dot-pkgs uses
			// the more explicit --no-use-air spelling.
			remaining = append(remaining, "--no-use-air")
		default:
			remaining = append(remaining, arg)
		}
	}
	return routePrefix, routePrefixSet, remaining, nil
}

func setRoutePrefixEnv(routePrefix string) (func(), error) {
	previous, existed := os.LookupEnv(routePrefixEnv)
	var err error
	if routePrefix == "" {
		err = os.Unsetenv(routePrefixEnv)
	} else {
		err = os.Setenv(routePrefixEnv, routePrefix)
	}
	if err != nil {
		return nil, err
	}
	return func() {
		if existed {
			_ = os.Setenv(routePrefixEnv, previous)
		} else {
			_ = os.Unsetenv(routePrefixEnv)
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
