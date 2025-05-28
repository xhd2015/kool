//go:build ignore
// +build ignore

package go_tools

import (
	"fmt"

	"github.com/xhd2015/less-gen/flags"
)

func handle(args []string) error {
	n := len(args)
	var remainArgs []string
	for i := 0; i < n; i++ {
		flag, value := flags.ParseIndex(args, &i)
		if flag == "" {
			remainArgs = append(remainArgs, args[i])
			continue
		}
		switch flag {
		case "-t", "--timeout":
			value, ok := value()
			if !ok {
				return fmt.Errorf("%s requires a value", flag)
			}
			_ = value
		// ...
		default:
			return fmt.Errorf("unknown flag: %s", flag)
		}
	}
	return nil
}
