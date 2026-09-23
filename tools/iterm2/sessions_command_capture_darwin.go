//go:build darwin

package iterm2

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/xhd2015/xgo/support/cmd"
	"golang.org/x/sys/unix"
)

func readForegroundIdentity(pid int) (foregroundProcessIdentity, error) {
	info, err := unix.SysctlKinfoProc("kern.proc.pid", pid)
	if err != nil {
		return foregroundProcessIdentity{}, err
	}
	return foregroundProcessIdentity{
		PID: int(info.Proc.P_pid), PPID: int(info.Eproc.Ppid),
		PGID: int(info.Eproc.Pgid), TPGID: int(info.Eproc.Tpgid), TTY: int(info.Eproc.Tdev),
		StartSec: info.Proc.P_starttime.Sec, StartUsec: int64(info.Proc.P_starttime.Usec),
	}, nil
}

func readForegroundArgv(pid int) ([]string, error) {
	// kern.procargs2 is KERN_PROCARGS2 (49), not the display-form ps command.
	data, err := unix.SysctlRaw("kern.procargs2", pid)
	if err != nil {
		return nil, err
	}
	defer clear(data) // discard the raw buffer, including any trailing environment
	return parseForegroundProcArgs(data)
}

func readForegroundCwd(pid int) (string, error) {
	// lsof field mode uses NUL delimiters so whitespace/newlines in cwd survive.
	// -b avoids blocking kernel calls; -S2 bounds lsof's kernel-call timeout.
	// Capture raw output rather than cmd.Output's trailing-newline trimming.
	var out bytes.Buffer
	err := cmd.New().Stdout(&out).Stderr(io.Discard).Run("/usr/sbin/lsof", "-b", "-S2", "-a", "-p", strconv.Itoa(pid), "-d", "cwd", "-Fn0")
	if err != nil {
		return "", err
	}
	for _, field := range bytes.Split(out.Bytes(), []byte{0}) {
		field = bytes.TrimPrefix(field, []byte{'\n'})
		if len(field) > 0 && field[0] == 'n' {
			return string(field[1:]), nil
		}
	}
	return "", fmt.Errorf("launcher cwd not found")
}
