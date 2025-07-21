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
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"os/exec"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/kool/pkgs/web"
)

// Re-enable embedded filesystem
//
//go:embed react/dist
var reactDistFS embed.FS

// Global variable to track PlantUML Docker container
var plantumlContainer struct {
	isRunning   bool
	port        int
	containerID string
}

type FileNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	IsDir    bool        `json:"isDir"`
	Children []*FileNode `json:"children,omitempty"`
}

// mimeTypeHandler wraps an http.Handler and sets proper MIME types
type mimeTypeHandler struct {
	handler http.Handler
}

func (h *mimeTypeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set MIME type based on file extension
	ext := filepath.Ext(r.URL.Path)
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	default:
		// Use Go's built-in MIME type detection for other files
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			w.Header().Set("Content-Type", mimeType)
		}
	}

	// Call the wrapped handler
	h.handler.ServeHTTP(w, r)
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

func Serve(dir string, plantumlServer string) error {
	return ServeWithInitialFile(dir, plantumlServer, "")
}

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

	// Serve static files from the embedded React build
	reactFileSystem, err := fs.Sub(reactDistFS, "react/dist")
	if err != nil {
		return fmt.Errorf("failed to create react file system: %v", err)
	}

	// Create sub-filesystem for assets
	assetsFileSystem, err := fs.Sub(reactFileSystem, "assets")
	if err != nil {
		return fmt.Errorf("failed to create assets file system: %v", err)
	}

	// Serve React assets from /assets/ path with proper MIME types
	http.Handle("/assets/", http.StripPrefix("/assets/", &mimeTypeHandler{http.FileServer(http.FS(assetsFileSystem))}))
	// Serve React static files like vite.svg from root
	http.Handle("/kool.svg", &mimeTypeHandler{http.FileServer(http.FS(reactFileSystem))})

	// Serve the main HTML page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		indexFile, err := reactFileSystem.Open("index.html")
		if err != nil {
			http.Error(w, "Failed to load index.html", http.StatusInternalServerError)
			return
		}
		defer indexFile.Close()

		content, err := io.ReadAll(indexFile)
		if err != nil {
			http.Error(w, "Failed to read index.html", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.Write(content)
		// http.Error(w, "React assets are not served directly from embedded FS. Please run with -embed flag.", http.StatusInternalServerError)
	})

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
			Path    string `json:"path"`
			Content string `json:"content"`
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

		// Find an available port
		port, err := web.FindAvailablePort(6743, 100)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to find available port: %v", err), http.StatusInternalServerError)
			return
		}

		// Create the Docker command with a name for easier management
		containerName := fmt.Sprintf("plantuml-server-%d", port)
		dockerCmd := fmt.Sprintf("docker run --rm --name %s -p %d:8080 plantuml/plantuml-server:jetty", containerName, port)

		// Update container tracking
		plantumlContainer.isRunning = true
		plantumlContainer.port = port
		plantumlContainer.containerID = containerName

		response := map[string]interface{}{
			"success": true,
			"port":    port,
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
