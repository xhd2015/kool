package cmd

import "github.com/xhd2015/xgo/support/cmd"

func Debug(b bool) *cmd.CmdBuilder {
	c := cmd.New()
	if b {
		c.Debug()
	}
	return c
}
