package vscodegit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ipcRetryCount = 3
	ipcRetryDelay = 100 * time.Millisecond
)

var ipcSocketPathOverride string

// SetIPC_SOCKETPathForTest overrides the IPC socket path for tests.
func SetIPC_SOCKETPathForTest(path string) {
	ipcSocketPathOverride = path
}

func ipcSocketPath() string {
	if ipcSocketPathOverride != "" {
		return ipcSocketPathOverride
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".kool", "xhd2015.open-in-new-window.sock")
	}
	return filepath.Join(home, ".kool", "xhd2015.open-in-new-window.sock")
}

type ipcResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

func sendIPC(op, path string, replace bool) error {
	socketPath := ipcSocketPath()
	var lastErr error
	for attempt := 0; attempt < ipcRetryCount; attempt++ {
		if attempt > 0 {
			time.Sleep(ipcRetryDelay)
		}
		err := tryIPC(socketPath, op, path, replace)
		if err == nil {
			return nil
		}
		lastErr = err
	}
	return lastErr
}

func tryIPC(socketPath, op, path string, replace bool) error {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	req := map[string]interface{}{"op": op, "path": path}
	if replace {
		req["replace"] = true
	}
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	if _, err := conn.Write(append(data, '\n')); err != nil {
		return err
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	var resp ipcResponse
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &resp); err != nil {
		return err
	}
	if !resp.OK {
		if resp.Error != "" {
			return fmt.Errorf("%s", resp.Error)
		}
		return fmt.Errorf("IPC request failed")
	}
	return nil
}