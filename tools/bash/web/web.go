package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/xhd2015/kool/tools/bash/history"
	_ "github.com/xhd2015/kool/tools/web"
	"github.com/xhd2015/kool/tools/web/server"
	"github.com/xhd2015/less-gen/flags"
)

const help = `
kool bash web starts a web interface for bash management.

Usage:
  kool bash web [OPTIONS]

Options:
  --port <port>    set the server port (default: 8080)
`

func Handle(args []string) error {
	var port int
	args, err := flags.
		Int("--port", &port).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %v", args)
	}

	html, err := server.FormatTemplateHtml(server.FormatOptions{
		Title:     "Bash History Manager",
		Component: "BashHistoryManager",
	})
	if err != nil {
		return err
	}

	return server.ServeComponent(port, server.ServeOptions{
		Static: server.StaticOptions{
			IndexHtml: html,
		},
		Route: func(mux *http.ServeMux) error {
			mux.HandleFunc("/api/bash/history", handleHistoryList)
			mux.HandleFunc("/api/bash/history/delete", handleHistoryDelete)
			return nil
		},
	})
}

func handleHistoryList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	page := 1
	pageSize := 20
	if p := r.URL.Query().Get("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if p := r.URL.Query().Get("pageSize"); p != "" {
		fmt.Sscanf(p, "%d", &pageSize)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	lines, err := history.GetHistoryLines()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get home history: %v", err), http.StatusInternalServerError)
		return
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filtered := make([]string, 0)
		for _, line := range lines {
			if strings.Contains(line, search) {
				filtered = append(filtered, line)
			}
		}
		lines = filtered
	}

	total := len(lines)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"list":  lines[start:end],
		"total": total,
	})
}

func handleHistoryDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Cmd string `json:"cmd"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Cmd == "" {
		http.Error(w, "cmd is required", http.StatusBadRequest)
		return
	}

	err := history.DelHistoryLine(req.Cmd)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete from history: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
