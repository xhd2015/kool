package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/xhd2015/kool/tools/bash/server/model"
	"github.com/xhd2015/less-gen/flags"
)

const help = `
kool bash serve starts an HTTP server for remote command execution.

Usage:
  kool bash serve --port PORT

Options:
  --port PORT    port to listen on (required)

Endpoints:
  POST /run      run a command and wait for completion
  POST /start    start a command and return immediately
  GET  /ps       list all running processes
  POST /kill     kill a specific process by PID
  POST /killall  kill all running processes
`

type Server struct {
	processes map[int]*model.ProcessInfo
	mutex     sync.RWMutex
}

func Handle(args []string) error {
	var port int
	args, err := flags.Int("--port", &port).Help("-h,--help", help).Parse(args)
	if err != nil {
		return err
	}

	if port == 0 {
		return fmt.Errorf("--port is required")
	}

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %v", args)
	}

	server := &Server{
		processes: make(map[int]*model.ProcessInfo),
	}

	return server.Start(port)
}

func (s *Server) Start(port int) error {
	http.HandleFunc("/run", s.handleRun)
	http.HandleFunc("/start", s.handleStart)
	http.HandleFunc("/ps", s.handlePS)
	http.HandleFunc("/kill", s.handleKill)
	http.HandleFunc("/killall", s.handleKillAll)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Starting bash server on %s\n", addr)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  POST %s/run     - run command and wait\n", addr)
	fmt.Printf("  POST %s/start   - start command and return PID\n", addr)
	fmt.Printf("  GET  %s/ps      - list running processes\n", addr)
	fmt.Printf("  POST %s/kill    - kill specific process by PID\n", addr)
	fmt.Printf("  POST %s/killall - kill all running processes\n", addr)

	return http.ListenAndServe(addr, nil)
}

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if req.Command == "" {
		http.Error(w, "Command is required", http.StatusBadRequest)
		return
	}

	cmd := exec.Command(req.Command, req.Args...)
	stdout, err := cmd.Output()

	var stderr []byte
	var exitCode int

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			stderr = exitError.Stderr
			exitCode = exitError.ExitCode()
		} else {
			// Command failed to start
			stderr = []byte(err.Error())
			exitCode = -1
		}
	}

	response := model.RunResponse{
		Stdout:   string(stdout),
		Stderr:   string(stderr),
		ExitCode: exitCode,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if req.Command == "" {
		http.Error(w, "Command is required", http.StatusBadRequest)
		return
	}

	// Build command string for comparison
	commandStr := req.Command
	if len(req.Args) > 0 {
		commandStr += " " + fmt.Sprintf("%v", req.Args)
	}

	// Check for singleton mode
	if req.Singleton {
		s.mutex.RLock()
		for _, proc := range s.processes {
			if proc.Command == commandStr {
				s.mutex.RUnlock()
				response := model.StartResponse{PID: proc.PID}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
				return
			}
		}
		s.mutex.RUnlock()
	}

	cmd := exec.Command(req.Command, req.Args...)

	// Start the command
	if err := cmd.Start(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to start command: %v", err), http.StatusInternalServerError)
		return
	}

	pid := cmd.Process.Pid

	// Store process info
	s.mutex.Lock()
	s.processes[pid] = &model.ProcessInfo{
		PID:     pid,
		Command: commandStr,
		Started: time.Now().Format(time.RFC3339),
	}
	s.mutex.Unlock()

	// Start a goroutine to clean up when process exits
	go func() {
		cmd.Wait()
		s.mutex.Lock()
		delete(s.processes, pid)
		s.mutex.Unlock()
	}()

	response := model.StartResponse{PID: pid}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handlePS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mutex.RLock()
	processes := make([]model.ProcessInfo, 0, len(s.processes))
	for _, proc := range s.processes {
		// Check if process is still running
		if s.isProcessRunning(proc.PID) {
			processes = append(processes, *proc)
		}
	}
	s.mutex.RUnlock()

	// Clean up dead processes
	s.mutex.Lock()
	for pid := range s.processes {
		if !s.isProcessRunning(pid) {
			delete(s.processes, pid)
		}
	}
	s.mutex.Unlock()

	response := model.PSResponse{Processes: processes}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleKill(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.KillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if req.PID <= 0 {
		http.Error(w, "Valid PID is required", http.StatusBadRequest)
		return
	}

	// Check if we're tracking this process
	s.mutex.RLock()
	_, exists := s.processes[req.PID]
	s.mutex.RUnlock()

	if !exists {
		response := model.KillResponse{
			Success: false,
			Message: fmt.Sprintf("Process %d not found in managed processes", req.PID),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Try to kill the process
	process, err := os.FindProcess(req.PID)
	if err != nil {
		response := model.KillResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to find process %d: %v", req.PID, err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	err = process.Kill()
	if err != nil {
		response := model.KillResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to kill process %d: %v", req.PID, err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Remove from our tracking
	s.mutex.Lock()
	delete(s.processes, req.PID)
	s.mutex.Unlock()

	response := model.KillResponse{
		Success: true,
		Message: fmt.Sprintf("Process %d killed successfully", req.PID),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleKillAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	var killedPIDs []int
	var errors []string

	for pid := range s.processes {
		process, err := os.FindProcess(pid)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to find process %d: %v", pid, err))
			continue
		}

		err = process.Kill()
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to kill process %d: %v", pid, err))
			continue
		}

		killedPIDs = append(killedPIDs, pid)
		delete(s.processes, pid)
	}

	response := model.KillAllResponse{
		KilledCount: len(killedPIDs),
		KilledPIDs:  killedPIDs,
	}

	if len(errors) > 0 {
		response.Errors = errors
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
