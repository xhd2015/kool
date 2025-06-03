package ai

import (
	"fmt"
	"strings"

	"github.com/xhd2015/less-gen/flags"
)

func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires sub commands: prompt")
	}
	cmd := args[0]
	args = args[1:]
	switch cmd {
	case "prompt":
		return handlePrompt(args)
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

const promptHelp = `
prompt help to dump directory structure and files

Usage: prompt <dir> [OPTIONS]

Options:
  -v,--verbose                     show verbose info

Examples:
  kool ai prompt ./sub_dir
`

func handlePrompt(args []string) error {
	n := len(args)
	var remainArgs []string
	var verbose bool
	var excludes []string
	for i := 0; i < n; i++ {
		flag, value := flags.ParseIndex(args, &i)
		if flag == "" {
			remainArgs = append(remainArgs, args[i])
			continue
		}
		_ = value
		switch flag {
		case "-v", "--verbose":
			verbose = true
		case "-h", "--help":
			fmt.Print(strings.TrimPrefix(promptHelp, "\n"))
			return nil
		case "--exclude":
			// exclude file patterns
			value, ok := value()
			if !ok {
				return fmt.Errorf("%s requires value", flag)
			}
			excludes = append(excludes, value)
		// ...
		default:
			return fmt.Errorf("unrecognized flag: %s", flag)
		}
	}

	_ = verbose
	_ = excludes
	if len(remainArgs) == 0 {
		return fmt.Errorf("requires dir")
	}
	if len(remainArgs) > 1 {
		return fmt.Errorf("requires only one dir, found extra: %s", strings.Join(remainArgs[1:], ","))
	}
	dir := remainArgs[0]
	return HandlePrompt(dir)
}
