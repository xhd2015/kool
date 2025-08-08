
### Go Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "time"
)

type Server struct {
    port string
}

func NewServer(port string) *Server {
    return &Server{port: port}
}

func (s *Server) Start(ctx context.Context) error {
    mux := http.NewServeMux()
    
    // Health check endpoint
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprint(w, "OK")
    })

    // API endpoints
    mux.HandleFunc("/api/users", s.handleUsers)
    
    server := &http.Server{
        Addr:    ":" + s.port,
        Handler: mux,
        ReadTimeout: 15 * time.Second,
        WriteTimeout: 15 * time.Second,
    }

    go func() {
        <-ctx.Done()
        server.Shutdown(context.Background())
    }()

    fmt.Printf("Server starting on port %s\n", s.port)
    return server.ListenAndServe()
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
    // Implementation here
}
```
