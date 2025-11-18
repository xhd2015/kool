package dlv

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/xhd2015/kool/tools/go/commands"
	"github.com/xhd2015/xgo/support/cmd"
)

const PROMPT_TEMPLATE = `
  > VSCode: add the following config to .vscode/launch.json configurations:
    {
        "configurations": [
                {
                        "name": "Debug dlv localhost:2345",
                        "type": "go",
                        "debugAdapter": "dlv-dap",
                        "request": "attach",
                        "mode": "remote",
                        "port": 2345,
                        "host": "127.0.0.1",
                        "cwd":"./"
                }
        ]
    }
    And set breakpoint at: __DEBUG_POINT__
  > GoLand: click Add Configuration > Go Remote > localhost:2345
  > Terminal: dlv connect localhost:2345
`

// dlv exec --listen=:2345 --api-version=2 --check-go-version=false --headless --
func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: dlv <program> [args...]")
	}
	n := len(args)
	var dlvFlags []string
	var remainArgs []string
	for i := 0; i < n; i++ {
		arg := args[i]
		if arg == "--" {
			remainArgs = append(remainArgs, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") {
			remainArgs = append(remainArgs, args[i:]...)
			break
		}
		dlvFlags = append(dlvFlags, arg)

		eqIdx := strings.Index(arg, "=")
		if eqIdx < 0 {
			if i+1 < n && !strings.HasPrefix(args[i+1], "-") {
				dlvFlags = append(dlvFlags, arg, args[i+1])
				i++
			}
		}
	}
	if len(remainArgs) == 0 {
		return fmt.Errorf("usage: dlv [flags...] <program> [args...]")
	}

	binary := remainArgs[0]
	remainArgs = remainArgs[1:]

	return Debug(binary, DebugOptions{
		Port:          2345,
		Args:          remainArgs,
		ExtraDlvFlags: dlvFlags,
	})
}

type DebugOptions struct {
	Dir           string
	Port          int
	PassStdin     bool
	Args          []string
	ExtraDlvFlags []string
}

func Debug(binary string, opts DebugOptions) error {
	dir := opts.Dir
	port := opts.Port
	dlvFlags := opts.ExtraDlvFlags
	args := opts.Args
	binPath, _ := exec.LookPath(binary)
	if binPath != "" {
		binary = binPath
	}

	if port == 0 {
		port = 2345
	}

	dlvArgs := []string{"exec",
		fmt.Sprintf("--listen=:%d", port),
		"--api-version=2",
		"--check-go-version=false",
		// "--tty=/dev/tty",
		"--headless",
	}
	dlvArgs = append(dlvArgs, dlvFlags...)
	dlvArgs = append(dlvArgs, "--")
	dlvArgs = append(dlvArgs, binary)
	dlvArgs = append(dlvArgs, args...)

	// let dlv print first
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Print(formatPrompt(binary))
	}()

	c := cmd.Debug()
	if opts.PassStdin {
		c.Stdin(os.Stdin)
	}
	return c.Dir(dir).Run("dlv", dlvArgs...)
}

func formatPrompt(binary string) string {
	prompt := PROMPT_TEMPLATE
	debugPoint := "main"

	mainFile := getBinaryMainFile(binary)
	if mainFile != "" {
		debugPoint = mainFile
	}
	if len(debugPoint) > 48 {
		debugPoint = "\n        " + debugPoint
	}

	prompt = strings.ReplaceAll(prompt, "__DEBUG_POINT__", debugPoint)
	return strings.TrimPrefix(prompt, "\n")
}

// example:
//
//	TEXT main.main(SB) /Users/xhd2015/Projects/xhd2015/kool/main.go
//	...
func objdump(binary string, symbol string) (string, error) {
	return cmd.Output("go", "tool", "objdump", "-s", symbol, binary)
}

func extractFileFromObjdump(objdump string) string {
	var firstLine string
	idx := strings.Index(objdump, "\n")
	if idx < 0 {
		firstLine = objdump
	} else {
		firstLine = objdump[:idx]
	}

	// main.main(SB) /Users/xhd2015/Projects/xhd2015/kool/main.go
	ANCHOR := ") "
	fnIdx := strings.LastIndex(firstLine, ANCHOR)
	if fnIdx < 0 {
		return ""
	}
	fnIdx += len(ANCHOR)

	return strings.TrimSpace(firstLine[fnIdx:])
}

func getBinaryMainFile(binary string) string {
	dump, err := objdump(binary, "main.main")
	if err != nil {
		return ""
	}
	return extractFileFromObjdump(dump)
}

func HasMainMain(binary string) bool {
	return getBinaryMainFile(binary) != ""
}

func getBinaryNM(binary string) string {
	// Run go tool nm to list symbols in the binary
	output, err := commands.GoToolNM(binary)
	if err != nil {
		return "" // fallback to a default
	}

	// Search for main.main symbol
	// output:
	//  10060efb0 T main.main
	//  10061fe90 T main.main.func1
	lines := strings.Split(string(output), "\n")
	var mainLine string
	for _, line := range lines {
		if strings.HasSuffix(line, " main.main") {
			mainLine = line
			break
		}
	}
	if mainLine != "" {
		// Once we find main.main, use addr2line to get file info
		fields := strings.Fields(mainLine)
		if len(fields) > 0 {
			// Instead of addr2line which might not be available on all platforms
			// Try using go tool objdump with grep
			objOutput, err := commands.GoToolObjdump("-s", "main.main", binary)
			if err == nil {
				objLines := strings.Split(objOutput, "\n")
				for _, objLine := range objLines {
					if strings.Contains(objLine, ".go:") {
						parts := strings.Split(objLine, ".go:")
						if len(parts) > 0 {
							file := parts[0]
							lastSlash := strings.LastIndex(file, "/")
							if lastSlash >= 0 {
								file = file[lastSlash+1:]
							}
							return file + ".go"
						}
					}
				}
			}
		}
		// If all else fails, return a default
		return ""
	}

	return ""
}
