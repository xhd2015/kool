package run

import (
	"fmt"
	"strings"

	"github.com/xhd2015/less-flags"
)

const help = `
Usage: __MODULE_NAME__ [OPTIONS]

Options:
  -h,--help            show help message
`

type Config struct{}

func Main(args []string) error {
	config := Config{}
	args, err := lessflags.
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}
	return Run(config)
}

func Run(config Config) error {
	return nil
}
