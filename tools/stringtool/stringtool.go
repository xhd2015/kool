package stringtool

import (
	"fmt"
	"strings"

	"github.com/xhd2015/kool/pkgs/terminal"
	"github.com/xhd2015/less-flags"
)

func HandleSplit(args []string) error {
	var separator string = ","
	args, err := lessflags.String("-s,--separator", &separator).
		Parse(args)
	if err != nil {
		return err
	}

	data, err := terminal.ReadOrTerminalData(args)
	if err != nil {
		return err
	}

	contents := strings.Split(data, ",")
	for _, content := range contents {
		s := strings.TrimSpace(content)
		if s == "" {
			continue
		}
		fmt.Println(s)
	}
	return nil
}
