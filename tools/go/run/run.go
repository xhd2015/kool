package run

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/kool/tools/dlv"
	"github.com/xhd2015/xgo/cmd/xgo/pathsum"
	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/netutil"
)

func Handle(args []string) error {
	var debug bool
	n := len(args)
	goArgs := make([]string, 0, n)
	var remainArgs []string

	var debugCwd string
	var hasGCflags bool
	for i := 0; i < n; i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			remainArgs = append(remainArgs, args[i:]...)
			break
		}
		if arg == "--debug" || arg == "-debug" {
			debug = true
			continue
		}

		if arg == "--debug-cwd" || arg == "-debug-cwd" || arg == "--debug-wd" || arg == "-debug-wd" {
			if i+1 >= n {
				return fmt.Errorf("%s requires argument", arg)
			}
			debugCwd = args[i+1]
			i++
			continue
		} else if suffix, ok := strings.CutPrefix(arg, "--debug-cwd="); ok {
			debugCwd = suffix
			continue
		} else if suffix, ok := strings.CutPrefix(arg, "-debug-cwd="); ok {
			debugCwd = suffix
			continue
		} else if suffix, ok := strings.CutPrefix(arg, "--debug-wd="); ok {
			debugCwd = suffix
			continue
		} else if suffix, ok := strings.CutPrefix(arg, "-debug-wd="); ok {
			debugCwd = suffix
			continue
		}

		if arg == "-gcflags=all=-N -l" || arg == "-gcflags=all=-l -N" {
			hasGCflags = true
		}
		goArgs = append(goArgs, arg)
		if !strings.Contains(arg, "=") {
			if i+1 < n && !strings.HasPrefix(args[i+1], "-") {
				goArgs = append(goArgs, args[i+1])
				i++
			}
		}
	}

	if !debug && debugCwd == "" {
		origArgs := []string{"run"}
		origArgs = append(origArgs, args...)
		return cmd.Debug().Run("go", origArgs...)
	}

	buildDir, err := getConsistentBuildDir()
	if err != nil {
		return err
	}
	debugBin := filepath.Join(buildDir, "__debug_bin")

	buildArgs := []string{
		"build",
	}
	buildArgs = append(buildArgs, goArgs...)
	if !hasGCflags {
		buildArgs = append(buildArgs, "-gcflags=all=-N -l")
	}
	buildArgs = append(buildArgs, "-o", debugBin)
	if len(remainArgs) > 0 {
		buildArgs = append(buildArgs, remainArgs[0])
		remainArgs = remainArgs[1:]
	}
	err = cmd.Debug().Run("go", buildArgs...)
	if err != nil {
		return err
	}
	if !debug {
		return cmd.Debug().Dir(debugCwd).Run(debugBin, remainArgs...)
	}
	return netutil.ServePort("localhost", 2345, true, 500*time.Millisecond, func(port int) {
		// fmt.Fprintln(os.Stdout, debug_util.FormatDlvPrompt(port))
	}, func(port int) error {

		// dlv exec --api-version=2 --listen=localhost:2345 --accept-multiclient --headless ./debug.bin
		return dlv.Debug(debugBin, dlv.DebugOptions{
			Dir:  debugCwd,
			Port: port,
			Args: remainArgs,
		})
	})
}

func getConsistentBuildDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	absWd, err := filepath.Abs(wd)
	if err != nil {
		return "", err
	}
	sum, err := pathsum.PathSum("go-build", absWd)
	if err != nil {
		return "", err
	}
	return filepath.Join(os.TempDir(), "kool-go-build", sum), nil
}
