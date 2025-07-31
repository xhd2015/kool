package model

type ProcessInfo struct {
	PID     int    `json:"pid"`
	Command string `json:"command"`
	Started string `json:"started"`
}

type RunRequest struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

type RunResponse struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

type StartRequest struct {
	Command   string   `json:"command"`
	Args      []string `json:"args,omitempty"`
	Singleton bool     `json:"singleton,omitempty"`
}

type StartResponse struct {
	PID int `json:"pid"`
}

type PSResponse struct {
	Processes []ProcessInfo `json:"processes"`
}

type KillRequest struct {
	PID int `json:"pid"`
}

type KillResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type KillAllResponse struct {
	KilledCount int      `json:"killed_count"`
	KilledPIDs  []int    `json:"killed_pids"`
	Errors      []string `json:"errors,omitempty"`
}
