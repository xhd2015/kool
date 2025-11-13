package run

import (
	"os"
	"path/filepath"
	"time"

	"github.com/xhd2015/kool/tools/dlv"
	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/xgo/cmd/xgo/pathsum"
	"github.com/xhd2015/xgo/support/cmd"
	"github.com/xhd2015/xgo/support/netutil"
)

const help = `
Run or debug go program

Usage: kool go run <cmd> [OPTIONS]

Available commands:
  create <name>                    create a new project
  help                             show help message

Options:
  --dir <dir>                      set the output directory
  -v,--verbose                     show verbose info  

Examples:
  kool go run help                         show help message
  kool go run create my_project            create a new project named my_project
`

func Handle(args []string) error {
	return HandleOpts(args, Options{
		AcceptDebugFlag: true,
	})
}

type Options struct {
	AcceptDebugFlag bool
	IsDebug         bool
}

const debugHelp = `
kool debug go run <cmd> [OPTIONS]

Debug options:
  --stdin                          pass stdin to the debugged program (default: false)
  --debug-cwd <dir>                set the debug working directory (default: current working directory)
  -debug,--debug                   enable debug mode (default: false)
  -debug-wd,--debug-wd             set the debug working directory (default: current working directory)
  -gcflags,--gcflags               set the gcflags (default: [])
  -h,--help                         show help message (default: false)

Examples:
  kool debug go run ./
  kool debug go run --stdin ./
`

func HandleOpts(args []string, opts Options) error {
	acceptDebugFlag := opts.AcceptDebugFlag
	isDebug := opts.IsDebug

	var debug bool

	var debugCwd string
	var gcflags []string
	var passStdin bool

	fb := flags.
		Bool("--stdin", &passStdin).
		StringSlice("-gcflags,--gcflags", &gcflags).
		Help("-h,--help", debugHelp)

	if acceptDebugFlag {
		fb.String("--debug-cwd", &debugCwd).
			Bool("-debug,--debug", &debug).
			String("-debug-wd,--debug-wd,-debug-cwd,--debug-cwd", &debugCwd)
	}

	remainArgs, err := fb.
		StopOnFirstArg().
		Parse(args)
	if err != nil {
		return err
	}

	debugMode := isDebug || debug

	if !debugMode && debugCwd == "" {
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

	var hasGCAllflags bool
	for _, gcflag := range gcflags {
		if debugMode && !hasGCAllflags && (gcflag == "all=-N -l" || gcflag == "all=-l -N") {
			hasGCAllflags = true
		}
		buildArgs = append(buildArgs, "-gcflags="+gcflag)
	}
	if debugMode && !hasGCAllflags {
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
	if !debugMode {
		return cmd.Debug().Dir(debugCwd).Run(debugBin, remainArgs...)
	}
	return DebugBinary(debugBin, remainArgs, DebugOptions{
		Cwd:       debugCwd,
		PassStdin: passStdin,
	})
}

type DebugOptions struct {
	Cwd       string
	PassStdin bool
}

func DebugBinary(binary string, args []string, opts DebugOptions) error {
	return netutil.ServePort("localhost", 2345, true, 500*time.Millisecond, func(port int) {
		// fmt.Fprintln(os.Stdout, debug_util.FormatDlvPrompt(port))
	}, func(port int) error {

		// dlv exec --api-version=2 --listen=localhost:2345 --accept-multiclient --headless ./debug.bin
		return dlv.Debug(binary, dlv.DebugOptions{
			Dir:       opts.Cwd,
			Port:      port,
			Args:      args,
			PassStdin: opts.PassStdin,
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
