package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/xhd2015/kool/tools/bash/server/model"
	"github.com/xhd2015/less-gen/flags"
)

const help = `
kool bash server-exec is a client for the bash server.

Usage:
  kool bash server-exec [OPTIONS] <command> [args...]

Commands:
  run <cmd> [args...]              run a command and wait for completion
  start <cmd> [args...]            start a command and return PID
  ps                               list all running processes
  kill <pid>                       kill a specific process by PID
  killall                          kill all running processes

Options:
  --server <url>                   server URL (default: http://localhost:8080)
  --singleton                      for start command: use singleton mode
  -h,--help                        show help message

Examples:
  kool bash server-exec run echo "Hello World"
  kool bash server-exec start sleep 60 --singleton
  kool bash server-exec ps
  kool bash server-exec kill 12345
  kool bash server-exec killall
  kool bash server-exec --server http://localhost:9090 ps
`

func Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires command: run, start, ps, kill, killall")
	}

	var serverURL string
	var singleton bool
	args, err := flags.String("--server", &serverURL).
		Bool("--singleton", &singleton).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	if len(args) == 0 {
		return fmt.Errorf("requires command: run, start, ps, kill, killall")
	}

	cmd := args[0]
	args = args[1:]

	switch cmd {
	case "run":
		return handleRun(serverURL, args)
	case "start":
		return handleStart(serverURL, args, singleton)
	case "ps":
		return handlePS(serverURL, args)
	case "kill":
		return handleKill(serverURL, args)
	case "killall":
		return handleKillAll(serverURL, args)
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func postJSON(url string, reqBody interface{}, respBody interface{}) error {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	if respBody != nil {
		err = json.NewDecoder(resp.Body).Decode(respBody)
		if err != nil {
			return fmt.Errorf("failed to decode response: %v", err)
		}
	}

	return nil
}

func getJSON(url string, respBody interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	if respBody != nil {
		err = json.NewDecoder(resp.Body).Decode(respBody)
		if err != nil {
			return fmt.Errorf("failed to decode response: %v", err)
		}
	}

	return nil
}

func handleRun(serverURL string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("run command requires at least one argument")
	}

	req := model.RunRequest{
		Command: args[0],
		Args:    args[1:],
	}

	var resp model.RunResponse
	err := postJSON(serverURL+"/run", req, &resp)
	if err != nil {
		return fmt.Errorf("failed to execute run command: %v", err)
	}

	if resp.Stdout != "" {
		fmt.Print(resp.Stdout)
	}
	if resp.Stderr != "" {
		fmt.Print(resp.Stderr)
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("command exited with code %d", resp.ExitCode)
	}

	return nil
}

func handleStart(serverURL string, args []string, singleton bool) error {
	if len(args) == 0 {
		return fmt.Errorf("start command requires at least one argument")
	}

	req := model.StartRequest{
		Command:   args[0],
		Args:      args[1:],
		Singleton: singleton,
	}

	var resp model.StartResponse
	err := postJSON(serverURL+"/start", req, &resp)
	if err != nil {
		return fmt.Errorf("failed to execute start command: %v", err)
	}

	fmt.Printf("Started process with PID: %d\n", resp.PID)
	return nil
}

func handlePS(serverURL string, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("ps command does not accept arguments")
	}

	var resp model.PSResponse
	err := getJSON(serverURL+"/ps", &resp)
	if err != nil {
		return fmt.Errorf("failed to execute ps command: %v", err)
	}

	if len(resp.Processes) == 0 {
		fmt.Println("No running processes")
		return nil
	}

	fmt.Printf("%-8s %-20s %s\n", "PID", "STARTED", "COMMAND")
	fmt.Printf("%-8s %-20s %s\n", "---", "-------", "-------")
	for _, proc := range resp.Processes {
		fmt.Printf("%-8d %-20s %s\n", proc.PID, proc.Started, proc.Command)
	}

	return nil
}

func handleKill(serverURL string, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("kill command requires exactly one PID argument")
	}

	pid, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid PID: %s", args[0])
	}

	req := model.KillRequest{
		PID: pid,
	}

	var resp model.KillResponse
	err = postJSON(serverURL+"/kill", req, &resp)
	if err != nil {
		return fmt.Errorf("failed to execute kill command: %v", err)
	}

	if resp.Success {
		fmt.Println(resp.Message)
	} else {
		return fmt.Errorf("kill failed: %s", resp.Message)
	}

	return nil
}

func handleKillAll(serverURL string, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("killall command does not accept arguments")
	}

	var resp model.KillAllResponse
	err := postJSON(serverURL+"/killall", map[string]interface{}{}, &resp)
	if err != nil {
		return fmt.Errorf("failed to execute killall command: %v", err)
	}

	if resp.KilledCount == 0 {
		fmt.Println("No processes to kill")
	} else {
		fmt.Printf("Killed %d process(es): %v\n", resp.KilledCount, resp.KilledPIDs)
	}

	if len(resp.Errors) > 0 {
		fmt.Println("Errors:")
		for _, errMsg := range resp.Errors {
			fmt.Printf("  %s\n", errMsg)
		}
	}

	return nil
}
