package grep

import (
	"fmt"
	"strings"

	"github.com/xhd2015/kool/tools/cmd"
	"github.com/xhd2015/less-gen/flags"
)

const help = `
Find some commit that has a file containing string "invalid param"

Usage: kool git grep <string>

git log -S "invalid param" --all --source -p
`

// find some commit that has a file containing string "invalid param":
// git log -S "invalid param" --all --source -p
func Handle(args []string) error {
	var verbose bool
	args, err := flags.Bool("-v,--verbose", &verbose).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("requires string")
	}
	search := args[0]
	args = args[1:]

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}

	return cmd.Debug(verbose).Run("git", "log", "-S", search, "--all", "--source", "-p")
}
