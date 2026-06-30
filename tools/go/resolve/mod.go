package resolve

import gotoolresolve "github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"

type ModuleInfo = gotoolresolve.ModuleInfo

type LocalModuleInfo = gotoolresolve.LocalModuleInfo

var (
	GetModuleInfo             = gotoolresolve.GetModuleInfo
	GetRootModulePath         = gotoolresolve.GetRootModulePath
	ResolveLocalModules       = gotoolresolve.ResolveLocalModules
	IsDependency              = gotoolresolve.IsDependency
	HasLocalFilesystemReplace = gotoolresolve.HasLocalFilesystemReplace
)