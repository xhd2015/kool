package with_go

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/kool/cmd/kool-with-go1.18/run"
)

func ResolveGoroot(goVersion string) (string, error) {
	switch goVersion {
	case "go1.25":
		goVersion = "go1.25.0"
	case "go1.24":
		goVersion = "go1.24.1"
	case "go1.23":
		goVersion = "go1.23.6"
	case "go1.22":
		goVersion = "go1.22.12"
	case "go1.21":
		goVersion = "go1.21.13"
	case "go1.20":
		goVersion = "go1.20.14"
	case "go1.19":
		goVersion = "go1.19.13"
	case "go1.18":
		goVersion = "go1.18.10"
	case "go1.17":
		goVersion = "go1.17.13"
	case "go1.16":
		goVersion = "go1.16.15"
	case "go1.15":
		goVersion = "go1.15.15"
	case "go1.14":
		goVersion = "go1.14.15"
	}
	return InstallGo(goVersion, "")
}

func ExecGoroot(goroot string, args []string, extraEnvs []string) error {
	absGoroot, err := filepath.Abs(goroot)
	if err != nil {
		return err
	}
	envs := os.Environ()
	envs = append(envs, extraEnvs...)
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
