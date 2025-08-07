package viewer

import (
	"context"
	"crypto/sha256"
	"embed"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"os/exec"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/xhd2015/kool/pkgs/web"
)

// isDockerRunningOnPort checks if a Docker container is running on the specified port
// by checking if the container with name plantuml-server-<port> is running
func isDockerRunningOnPort(port int) bool {
	containerName := fmt.Sprintf("plantuml-server-%d", port)

	// Use docker ps to check if container is running
	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", containerName), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		// Docker command failed, assume container is not running
		return false
	}

	// Check if the container name appears in the output
	outputStr := strings.TrimSpace(string(output))
	return outputStr == containerName
}

// Re-enable embedded filesystem
// prefix with all to include files starts with _.
// see https://pkg.go.dev/embed
//
//go:embed all:react/dist
var reactDistFS embed.FS

// Global variable to track PlantUML Docker container
var plantumlContainer struct {
	isRunning   bool
	port        int
	containerID string
}

// Global file watcher
var fileWatcher struct {
	watcher *fsnotify.Watcher

	mutex   sync.Mutex
	clients map[*websocket.Conn]bool
}

type FileNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	IsDir    bool        `json:"isDir"`
	Children []*FileNode `json:"children,omitempty"`
}

var plantUMLCacheDir string

func init() {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		// Fallback to temp directory if user cache dir is not available
		plantUMLCacheDir = filepath.Join(os.TempDir(), "plant-uml", "svg")
	} else {
		plantUMLCacheDir = filepath.Join(userCacheDir, "plant-uml", "svg")
	}
}

// generateGitDiff creates a git diff between old and new content
func generateGitDiff(oldContent, newContent, filename string) (string, error) {
	// Create temporary directory for git diff
	tempDir, err := os.MkdirTemp("", "git-diff-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create temp files
	oldFile := filepath.Join(tempDir, "old_"+filename)
	newFile := filepath.Join(tempDir, "new_"+filename)

	if err := os.WriteFile(oldFile, []byte(oldContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write old file: %v", err)
	}

	if err := os.WriteFile(newFile, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write new file: %v", err)
	}

	// Run git diff
	cmd := exec.Command("git", "diff", "--no-index", "--no-prefix", oldFile, newFile)
	output, err := cmd.Output()

	// git diff returns exit code 1 when files differ, which is expected
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// This is expected when files differ
		} else {
			return "", fmt.Errorf("git diff failed: %v", err)
		}
	}

	return string(output), nil
}

func Serve(dir string, plantumlServer string) error {
	return ServeWithInitialFile(dir, plantumlServer, "")
}

const PLANT_UTML_PORT = 6743

func ServeWithInitialFile(dir string, plantumlServer string, initialFile string) error {
	// Convert to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(plantUMLCacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", absDir)
	}

	port, err := web.FindAvailablePort(8080, 100)
	if err != nil {
		return err
	}

	err = Static(http.DefaultServeMux)
	if err != nil {
		return err
	}
	// Check if PlantUML server is already running
	if isDockerRunningOnPort(PLANT_UTML_PORT) {
		plantumlContainer.isRunning = true
		plantumlContainer.port = PLANT_UTML_PORT
		plantumlContainer.containerID = fmt.Sprintf("plantuml-server-%d", PLANT_UTML_PORT)
	}

	// API to get directory tree
	http.HandleFunc("/api/tree", func(w http.ResponseWriter, r *http.Request) {
		tree, err := buildFileTreeWithRelativePaths(absDir, absDir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tree)
	})

	// API to serve cached PlantUML SVG images
	http.HandleFunc("/planuml/svg/", func(w http.ResponseWriter, r *http.Request) {
		// Extract encoded string from URL path
		path := strings.TrimPrefix(r.URL.Path, "/planuml/svg/")
		if path == "" {
			http.Error(w, "encoded parameter is required", http.StatusBadRequest)
			return
		}

		// Generate cache file path using SHA256 hash of the encoded string for safety
		hash := sha256.Sum256([]byte(path))
		cacheFileName := hex.EncodeToString(hash[:]) + ".svg"
		cacheFilePath := filepath.Join(plantUMLCacheDir, cacheFileName)

		// Check if file exists in cache
		if _, err := os.Stat(cacheFilePath); err == nil {
			// Serve from cache
			http.ServeFile(w, r, cacheFilePath)
			return
		}

		plantServer := strings.TrimSuffix(plantumlServer, "/")

		// If we have a running local PlantUML container, use that instead
		if plantumlContainer.isRunning && plantumlContainer.port > 0 {
			plantServer = fmt.Sprintf("http://localhost:%d", plantumlContainer.port)
		} else if plantServer == "" {
			plantServer = "https://www.plantuml.com/plantuml"
		}

		// File not in cache, fetch from PlantUML service
		plantUMLURL := fmt.Sprintf("%s/svg/%s", plantServer, path)

		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		resp, err := client.Get(plantUMLURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to fetch from PlantUML: %v", err), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, fmt.Sprintf("PlantUML service returned status: %d", resp.StatusCode), http.StatusBadGateway)
			return
		}

		// Read the response body
		svgData, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read PlantUML response: %v", err), http.StatusInternalServerError)
			return
		}

		// Only cache if we have valid SVG content (basic validation)
		if len(svgData) > 0 && strings.Contains(string(svgData), "<svg") {
			// Save to cache
			if err := os.WriteFile(cacheFilePath, svgData, 0644); err != nil {
				// Log error but don't fail the request
				fmt.Printf("Warning: Failed to cache PlantUML SVG: %v\n", err)
			}
		}

		// Serve the SVG content
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours
		w.Write(svgData)
	})

	// API to get file content for preview
	http.HandleFunc("/api/preview", func(w http.ResponseWriter, r *http.Request) {
		filePath := r.URL.Query().Get("path")
		if filePath == "" {
			http.Error(w, "path parameter is required", http.StatusBadRequest)
			return
		}

		// Convert relative path to absolute path by joining with base directory
		var absFilePath string
		if filepath.IsAbs(filePath) {
			// If it's already absolute, use it directly (for backward compatibility)
			absFilePath = filePath
		} else {
			// If it's relative, join with base directory
			absFilePath = filepath.Join(absDir, filePath)
		}

		// Clean the path to resolve any ".." components
		absFilePath = filepath.Clean(absFilePath)

		// Security check: ensure the file is within the served directory
		if !strings.HasPrefix(absFilePath, absDir) {
			http.Error(w, "access denied", http.StatusForbidden)
			return
		}

		// Check if file exists
		stat, err := os.Stat(absFilePath)
		if os.IsNotExist(err) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		if stat.IsDir() {
			http.Error(w, "cannot preview directory", http.StatusBadRequest)
			return
		}

		ext := strings.ToLower(filepath.Ext(absFilePath))

		// Handle UML files
		if ext == ".uml" || ext == ".puml" {
			content, err := os.ReadFile(absFilePath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			response := map[string]interface{}{
				"type":    "uml",
				"content": string(content),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Handle Mermaid files
		if ext == ".mmd" {
			content, err := os.ReadFile(absFilePath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			response := map[string]interface{}{
				"type":    "mermaid",
				"content": string(content),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Handle Markdown files
		if ext == ".md" {
			content, err := os.ReadFile(absFilePath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			response := map[string]interface{}{
				"type":    "markdown",
				"content": string(content),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// For other files, return as text (could be extended for other types)
		content, err := os.ReadFile(absFilePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"type":    "text",
			"content": string(content),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Initialize file watcher
	if err := initFileWatcher(absDir); err != nil {
		return fmt.Errorf("failed to initialize file watcher: %v", err)
	}

	// API to handle file change notifications via WebSocket
	http.HandleFunc("/api/file-changes", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("Failed to upgrade to WebSocket for file changes: %v\n", err)
			return
		}
		defer func() {
			// Remove client when connection closes
			fileWatcher.mutex.Lock()
			if fileWatcher.clients != nil {
				delete(fileWatcher.clients, ws)
			}
			fileWatcher.mutex.Unlock()
			ws.Close()
		}()

		// Add client to the list
		fileWatcher.mutex.Lock()
		if fileWatcher.clients == nil {
			fileWatcher.clients = make(map[*websocket.Conn]bool)
		}
		fileWatcher.clients[ws] = true
		fileWatcher.mutex.Unlock()

		fmt.Println("File changes WebSocket connection established")

		// Keep the connection alive and handle ping messages
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				fmt.Printf("File changes WebSocket read error: %v\n", err)
				break
			}
		}
	})

	// API to execute terminal commands via WebSocket streaming
	http.HandleFunc("/api/terminal/stream", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("WebSocket connection request received")

		// Upgrade HTTP connection to WebSocket
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all connections for now
			},
		}

		// Check if this is a WebSocket upgrade request
		if r.Header.Get("Upgrade") != "websocket" {
			fmt.Printf("Not a WebSocket upgrade request. Headers: %+v\n", r.Header)
			http.Error(w, "Expected WebSocket upgrade", http.StatusBadRequest)
			return
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("Failed to upgrade to WebSocket: %v\n", err)
			return
		}
		defer ws.Close()
		fmt.Println("WebSocket connection established")

		// Get or create bash session
		session, err := initBashSession(absDir)
		if err != nil {
			fmt.Printf("Failed to create bash session: %v\n", err)
			ws.WriteJSON(map[string]string{"error": "Failed to create bash session: " + err.Error()})
			return
		}
		fmt.Println("Bash session obtained")

		// Create a context for this connection
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		// Handle incoming messages from WebSocket (user input)
		go func() {
			defer cancel()
			for {
				var msg map[string]interface{}
				if err := ws.ReadJSON(&msg); err != nil {
					fmt.Printf("Error reading WebSocket message: %v\n", err)
					return
				}

				fmt.Printf("Received WebSocket message: %+v\n", msg)

				if input, ok := msg["input"].(string); ok {
					fmt.Printf("Sending input to bash: %q\n", input)
					if err := session.sendInput(input); err != nil {
						fmt.Printf("Error sending input to bash: %v\n", err)
						return
					}
				}

				if resize, ok := msg["resize"].(map[string]interface{}); ok {
					if cols, colsOk := resize["cols"].(float64); colsOk {
						if rows, rowsOk := resize["rows"].(float64); rowsOk {
							fmt.Printf("Received resize request: cols=%d, rows=%d\n", int(cols), int(rows))
							if err := session.setSize(int(cols), int(rows)); err != nil {
								fmt.Printf("Error setting terminal size: %v\n", err)
							}
						}
					}
				}
			}
		}()

		fmt.Println("Starting output loop")
		// Listen for output from the bash session
		for {
			select {
			case output := <-session.outputChannel:
				fmt.Printf("Bash output: %q\n", output)
				if err := ws.WriteJSON(map[string]string{"output": output}); err != nil {
					fmt.Printf("Error writing output to WebSocket: %v\n", err)
					return
				}
			case errOutput := <-session.errorChannel:
				fmt.Printf("Bash error: %q\n", errOutput)
				if err := ws.WriteJSON(map[string]string{"error": errOutput}); err != nil {
					fmt.Printf("Error writing error to WebSocket: %v\n", err)
					return
				}
			case <-ctx.Done():
				fmt.Println("WebSocket context cancelled")
				return
			case <-time.After(30 * time.Second):
				fmt.Println("Sending keepalive")
				// Send keepalive
				if err := ws.WriteJSON(map[string]bool{"keepalive": true}); err != nil {
					fmt.Printf("Error sending keepalive: %v\n", err)
					return
				}
			}
		}
	})

	// API to get file content for editing
	http.HandleFunc("/api/content", func(w http.ResponseWriter, r *http.Request) {
		filePath := r.URL.Query().Get("path")
		if filePath == "" {
			http.Error(w, "path parameter is required", http.StatusBadRequest)
			return
		}

		// Convert relative path to absolute path by joining with base directory
		var absFilePath string
		if filepath.IsAbs(filePath) {
			absFilePath = filePath
		} else {
			absFilePath = filepath.Join(absDir, filePath)
		}

		// Clean the path to resolve any ".." components
		absFilePath = filepath.Clean(absFilePath)

		// Security check: ensure the file is within the served directory
		if !strings.HasPrefix(absFilePath, absDir) {
			http.Error(w, "access denied", http.StatusForbidden)
			return
		}

		// Check if file exists
		stat, err := os.Stat(absFilePath)
		if os.IsNotExist(err) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		if stat.IsDir() {
			http.Error(w, "cannot load directory content", http.StatusBadRequest)
			return
		}

		// Read file content
		content, err := os.ReadFile(absFilePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"content": string(content),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// API to save file content
	http.HandleFunc("/api/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var requestBody struct {
			Path       string `json:"path"`
			Content    string `json:"content"`
			OldContent string `json:"oldContent"`
		}

		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if requestBody.Path == "" {
			http.Error(w, "path is required", http.StatusBadRequest)
			return
		}

		// Convert relative path to absolute path by joining with base directory
		var absFilePath string
		if filepath.IsAbs(requestBody.Path) {
			absFilePath = requestBody.Path
		} else {
			absFilePath = filepath.Join(absDir, requestBody.Path)
		}

		// Clean the path to resolve any ".." components
		absFilePath = filepath.Clean(absFilePath)

		// Security check: ensure the file is within the served directory
		if !strings.HasPrefix(absFilePath, absDir) {
			http.Error(w, "access denied", http.StatusForbidden)
			return
		}

		// Read current file content to check for conflicts
		currentContent, err := os.ReadFile(absFilePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read current file content: %v", err), http.StatusInternalServerError)
			return
		}

		// Check if the old content matches the current file content
		if string(currentContent) != requestBody.OldContent {
			// Conflict detected - generate diffs
			filename := filepath.Base(requestBody.Path)

			// Generate diff between user's changes and original
			userDiff, err := generateGitDiff(requestBody.OldContent, requestBody.Content, filename)
			if err != nil {
				userDiff = fmt.Sprintf("Error generating user diff: %v", err)
			}

			// Generate diff between original and current file
			currentDiff, err := generateGitDiff(requestBody.OldContent, string(currentContent), filename)
			if err != nil {
				currentDiff = fmt.Sprintf("Error generating current diff: %v", err)
			}

			response := map[string]interface{}{
				"success":        false,
				"conflict":       true,
				"currentContent": string(currentContent),
				"userDiff":       userDiff,
				"currentDiff":    currentDiff,
				"message":        "File has been modified by another process. Please reload and try again.",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Write file content
		if err := os.WriteFile(absFilePath, []byte(requestBody.Content), 0644); err != nil {
			http.Error(w, fmt.Sprintf("failed to save file: %v", err), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"success": true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// API to start PlantUML Docker server
	http.HandleFunc("/api/start-plantuml", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// First, check if port 6743 is already running a Docker container
		if isDockerRunningOnPort(PLANT_UTML_PORT) {
			// Port 6743 is already running, just update our tracking
			plantumlContainer.isRunning = true
			plantumlContainer.port = PLANT_UTML_PORT
			plantumlContainer.containerID = fmt.Sprintf("plantuml-server-%d", PLANT_UTML_PORT) // Assume standard naming

			response := map[string]interface{}{
				"success": true,
				"port":    PLANT_UTML_PORT,
				// Don't send command since container is already running
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Create the Docker command with a name for easier management
		containerName := fmt.Sprintf("plantuml-server-%d", PLANT_UTML_PORT)
		dockerCmd := fmt.Sprintf("docker run --rm --name %s -p %d:8080 plantuml/plantuml-server:jetty", containerName, PLANT_UTML_PORT)

		// Update container tracking
		plantumlContainer.isRunning = true
		plantumlContainer.port = PLANT_UTML_PORT
		plantumlContainer.containerID = containerName

		response := map[string]interface{}{
			"success": true,
			"port":    PLANT_UTML_PORT,
			"command": dockerCmd,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// API to stop PlantUML Docker server
	http.HandleFunc("/api/stop-plantuml", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check if container is actually running
		if !plantumlContainer.isRunning || plantumlContainer.containerID == "" {
			http.Error(w, "no PlantUML container is running", http.StatusBadRequest)
			return
		}

		// Execute the Docker stop command directly on the backend
		cmd := exec.Command("docker", "stop", plantumlContainer.containerID)
		err := cmd.Run()

		// Update container tracking (regardless of command success to avoid inconsistent state)
		plantumlContainer.isRunning = false
		plantumlContainer.port = 0
		containerID := plantumlContainer.containerID
		plantumlContainer.containerID = ""

		if err != nil {
			// Log the error but still return success to frontend since we updated the state
			fmt.Printf("Warning: Failed to stop Docker container %s: %v\n", containerID, err)
		} else {
			fmt.Printf("Successfully stopped Docker container %s\n", containerID)
		}

		response := map[string]interface{}{
			"success": true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// API to check PlantUML Docker server status
	http.HandleFunc("/api/plantuml-status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		response := map[string]interface{}{
			"isRunning": plantumlContainer.isRunning,
			"port":      plantumlContainer.port,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Remove the old /api/terminal/input endpoint since we're using WebSocket

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	fmt.Printf("Serving directory preview at http://localhost:%d\n", port)
	fmt.Printf("Directory: %s\n", absDir)

	go func() {
		time.Sleep(1 * time.Second)
		web.OpenBrowser(fmt.Sprintf("http://localhost:%d", port))
	}()

	return server.ListenAndServe()
}

func buildFileTreeWithRelativePaths(rootPath, baseDir string) (*FileNode, error) {
	stat, err := os.Stat(rootPath)
	if err != nil {
		return nil, err
	}

	// Generate relative path from baseDir
	relativePath, err := filepath.Rel(baseDir, rootPath)
	if err != nil {
		return nil, err
	}

	// If it's the root directory, use "." as the relative path
	if relativePath == "." {
		relativePath = filepath.Base(rootPath)
	}

	node := &FileNode{
		Name:  filepath.Base(rootPath),
		Path:  relativePath,
		IsDir: stat.IsDir(),
	}

	if stat.IsDir() {
		entries, err := os.ReadDir(rootPath)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			// Skip hidden files and directories
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			childPath := filepath.Join(rootPath, entry.Name())
			childNode, err := buildFileTreeWithRelativePaths(childPath, baseDir)
			if err != nil {
				// Skip files that can't be read
				continue
			}
			node.Children = append(node.Children, childNode)
		}
	}

	return node, nil
}

// initFileWatcher initializes fsnotify watcher for the given directory
func initFileWatcher(dir string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %v", err)
	}

	fileWatcher.watcher = watcher
	fileWatcher.mutex.Lock()
	fileWatcher.clients = make(map[*websocket.Conn]bool)
	fileWatcher.mutex.Unlock()

	// Add the root directory to watch
	err = watcher.Add(dir)
	if err != nil {
		return fmt.Errorf("failed to watch directory %s: %v", dir, err)
	}

	// Walk through subdirectories and add them to watch
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files/dirs we can't access
		}

		// Skip hidden directories and files
		if strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				fmt.Printf("Warning: failed to watch directory %s: %v\n", path, err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory tree: %v", err)
	}

	// Start watching for events
	go watchFileEvents(dir)

	return nil
}

// watchFileEvents handles file system events and notifies clients
func watchFileEvents(baseDir string) {
	for {
		select {
		case event, ok := <-fileWatcher.watcher.Events:
			if !ok {
				return
			}

			// Skip hidden files and directories
			if strings.HasPrefix(filepath.Base(event.Name), ".") {
				continue
			}

			fmt.Printf("File event: %s - %s\n", event.Op, event.Name)

			// Handle different event types
			var eventType string
			var needsTreeRefresh bool

			if event.Op&fsnotify.Create == fsnotify.Create {
				eventType = "create"
				needsTreeRefresh = true

				// If a new directory is created, add it to the watcher
				if stat, err := os.Stat(event.Name); err == nil && stat.IsDir() {
					fileWatcher.watcher.Add(event.Name)
				}
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				eventType = "delete"
				needsTreeRefresh = true
			} else if event.Op&fsnotify.Write == fsnotify.Write {
				eventType = "modify"
			} else if event.Op&fsnotify.Rename == fsnotify.Rename {
				eventType = "rename"
				needsTreeRefresh = true
			} else {
				continue // Skip other events
			}

			// Get relative path from base directory
			relativePath, err := filepath.Rel(baseDir, event.Name)
			if err != nil {
				relativePath = event.Name
			}

			// Notify all connected clients
			notification := map[string]interface{}{
				"type":             "file_change",
				"event":            eventType,
				"path":             relativePath,
				"needsTreeRefresh": needsTreeRefresh,
			}

			notifyClients(notification)

		case err, ok := <-fileWatcher.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("File watcher error: %v\n", err)
		}
	}
}

// notifyClients sends a notification to all connected WebSocket clients
func notifyClients(notification map[string]interface{}) {
	fileWatcher.mutex.Lock()
	defer fileWatcher.mutex.Unlock()

	if fileWatcher.clients == nil {
		return
	}

	// Send to all connected clients
	for client := range fileWatcher.clients {
		err := client.WriteJSON(notification)
		if err != nil {
			fmt.Printf("Error sending notification to client: %v\n", err)
			// Remove client if send failed
			delete(fileWatcher.clients, client)
			client.Close()
		}
	}
}
