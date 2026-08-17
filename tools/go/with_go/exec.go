package with_go

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/withgo"
	"github.com/xhd2015/kool/tools/go/run"
)

func ResolveGoroot(goVersion string) (string, error) {
	return ResolveGorootWith(goVersion, withgo.ResolveOptions{Download: true})
}

func ResolveGorootWith(goVersion string, opts withgo.ResolveOptions) (string, error) {
	if opts.InstallDir == "" {
		dir, err := withgo.DefaultInstallDir()
		if err != nil {
			return "", err
		}
		opts.InstallDir = dir
	}
	return withgo.ResolveGoroot(goVersion, opts)
}

func ExecGoroot(goroot string, args []string, extraEnvs []string) error {
	absGoroot, err := filepath.Abs(goroot)
	if err != nil {
		return err
	}

	argEnvs, args := takeEnvs(args)

	envs := os.Environ()
	envs = append(envs, extraEnvs...)
	envs = append(envs, argEnvs...)

	PATH := filepath.Join(absGoroot, "bin") + string(os.PathListSeparator) + os.Getenv("PATH")
	envs = append(envs, "GOROOT="+absGoroot, "PATH="+PATH)

	if len(args) >= 2 && args[0] == "go" && args[1] == "run" {
		err := os.Setenv("PATH", PATH)
		if err != nil {
			return err
		}
		err = os.Setenv("GOROOT", absGoroot)
		if err != nil {
			return err
		}

		// use kool go
		return run.Handle(args[2:])
	}

	var targetCmd string
	var targetArgs []string
	if len(args) == 0 {
		targetCmd = "env"
	} else {
		targetCmd = args[0]
		targetArgs = args[1:]

		strip := strings.TrimPrefix(targetCmd, "./")
		if strip == filepath.Base(targetCmd) {
			// try lookup in $GOROOT/bin
			fullCmd := filepath.Join(absGoroot, "bin", targetCmd)
			if _, err := os.Stat(fullCmd); err == nil {
				targetCmd = fullCmd
			}
		}
	}

	execCmd := exec.Command(targetCmd, targetArgs...)
	execCmd.Env = envs
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	return execCmd.Run()
}

func takeEnvs(rawArgs []string) (envs []string, args []string) {
	i := 0
	n := len(rawArgs)
	for i < n {
		arg := rawArgs[i]
		if !strings.Contains(arg, "=") {
			args = rawArgs[i:]
			break
		}
		envs = append(envs, arg)
		i++
	}
	return
}
