package commands

import gotoolcommands "github.com/xhd2015/dot-pkgs/go-pkgs/gotool/commands"

type GoMod = gotoolcommands.GoMod

type GoModEditOptions = gotoolcommands.GoModEditOptions

var DefaultGoModEditOptions = gotoolcommands.DefaultGoModEditOptions

var (
	GoModEditJSON     = gotoolcommands.GoModEditJSON
	GoModEditReplace  = gotoolcommands.GoModEditReplace
	GoModDropReplace  = gotoolcommands.GoModDropReplace
	GoModEditRequire  = gotoolcommands.GoModEditRequire
	GoModTidy         = gotoolcommands.GoModTidy
)