# Development Guidelines

This document describes how to develop this Go server project.

## Project Structure

```
server_go_template/
├── main.go              # Entry point
├── config/              # Configuration loading and management
├── dao/                 # Database access layer
│   ├── dao.go           # DAO initialization
│   ├── engine/          # Database engine singleton
│   └── user/            # User-related tables
│       └── t_user/      # Table definition for t_user
├── env/                 # Environment variables
├── handle/              # HTTP handlers (gin handlers)
├── lib/                 # Reusable libraries
│   ├── log/             # Structured logging (slog + zap)
│   ├── metrics/         # Prometheus metrics
│   ├── middleware/      # HTTP middlewares
│   ├── routehelp/       # Route helpers (abort, success, parse)
│   ├── routewrap/       # Type-safe route wrapper
│   ├── server_errors/   # Application-level errors
│   └── trace/           # Request tracing
├── model/               # Business models (DTOs, request/response types)
├── route/               # Route registration
│   └── processor/       # Request processor
├── service/             # Business logic layer
├── task/                # Background tasks and cron jobs
└── types/               # Shared type definitions (IDs, enums)
```

## Package Responsibilities

### `dao/` - Database Access Layer

Define database tables in `dao/<domain>/t_<table_name>/` packages:

```go
// dao/user/t_user/t_user.go
package t_user

import (
    "github.com/xhd2015/arc-orm/orm"
    "github.com/xhd2015/arc-orm/table"
)

// SQL DDL as comment for reference
/*
CREATE TABLE `t_user` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(128) NOT NULL DEFAULT '',
    PRIMARY KEY (`id`)
);
*/

var Table = table.New("t_user")

var (
    ID   = Table.Int64("id")
    Name = Table.String("name")
)

// ORM binding for type-safe queries
var ORM = orm.Bind[User, UserOptional](engine.Engine, Table)

// Row struct - matches database columns
type User struct {
    Id   types.UserID `json:"id"`
    Name string       `json:"name"`
}

// Optional struct - for partial updates
type UserOptional struct {
    Id   *types.UserID `json:"id"`
    Name *string       `json:"name"`
}
```

### `types/` - Shared Type Definitions

Define ID types and enums in `types/types.go`:

```go
package types

// Base ID type
type ID int64

// Domain-specific IDs for type safety
type UserID int64
type OrderID int64
```

### `lib/server_errors/` - Application Errors

Define application-level errors in `lib/server_errors/errors.go`:

```go
package server_errors

import "errors"

var (
    ErrUserNotFound    = errors.New("user not found")
    ErrInvalidPassword = errors.New("invalid password")
    ErrTokenExpired    = errors.New("token expired")
)
```

### `model/` - Business Models

Define DTOs, request/response types in `model/`:

```go
// model/user.go
package model

type LoginRequest struct {
    Name string `json:"name"`
    X    string `json:"x"` // base64 encoded password
}

type LoginResponse struct {
    Code int    `json:"code"`
    Msg  string `json:"msg,omitempty"`
}
```

### `service/` - Business Logic

Implement business logic in `service/<domain>/`:

```go
// service/user/user.go
package user

func GetUserByID(ctx context.Context, userID types.UserID) (*t_user.User, error) {
    user, err := t_user.ORM.SelectAll().Where(t_user.ID.Eq(int64(userID))).QueryOne(ctx)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, server_errors.ErrUserNotFound
    }
    return user, nil
}
```

### `handle/` - HTTP Handlers

Implement HTTP handlers in `handle/<domain>/`:

```go
// handle/example/example.go
package example

type GetRequest struct {
    ID int64 `form:"id"`
}

type GetResponse struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}

func Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
    // Business logic here
    return &GetResponse{ID: req.ID, Name: "Example"}, nil
}
```

### `route/` - Route Registration

Register routes in `route/route.go`:

```go
package route

func Init(r *gin.Engine) {
    r.GET("/api/example", processor.Gin(example.Get))
    r.POST("/api/example/create", processor.Gin(example.Create))
}
```

## Logging

The logging library uses `slog` with `zap` backend for high-performance structured logging.
Trace IDs and caller info (file:line) are automatically included in logs.

### Basic Usage

```go
import "github.com/xhd2015/kool/tools/create/server_go_db_template/lib/log"

func MyHandler(ctx context.Context) {
    log.Info(ctx, "processing request")
    log.Infof(ctx, "processing request", "request_id", requestID, "user_id", userID)
    log.Errorf(ctx, "failed to process", "error", err)
}
```

### Structured Logging

Use key-value pairs for structured logging:

```go
log.Infof(ctx, "user data", "user_id", 1, "name", "John")
// Output: {"level":"INFO","time":"...","msg":"user data","caller":"handler.go:42","trace_id":"...","user_id":1,"name":"John"}
```

### Available Functions

- `log.Infof(ctx, msg, args...)` - Info level with key-value pairs
- `log.Errorf(ctx, msg, args...)` - Error level with key-value pairs
- `log.Warnf(ctx, msg, args...)` - Warn level with key-value pairs
- `log.Debugf(ctx, msg, args...)` - Debug level with key-value pairs
- `log.Info(ctx, msg)` - Info level without extra fields
- `log.Error(ctx, msg)` - Error level without extra fields

### Log Configuration

Configure via command line flags or config file:

- `--log-path` - Log file path (empty for stderr only)
- `--log-level` - Log level: debug, info, warn, error
- `--log-max-size` - Max log file size in MB before rotation
- `--log-max-backups` - Max number of old log files to keep
- `--log-max-age` - Max days to retain old log files
- `--log-compress` - Compress rotated log files

## Database Operations

Use arc-orm for type-safe database operations.

### Query

```go
// Single record
user, err := t_user.ORM.SelectAll().Where(t_user.ID.Eq(123)).QueryOne(ctx)

// Multiple records
users, err := t_user.ORM.SelectAll().Where(t_user.Name.Like("%test%")).Query(ctx)

// Count
count, err := t_user.ORM.Count().Where(t_user.Name.Eq("test")).Query(ctx)
```

### Insert

```go
id, err := t_user.ORM.Insert(ctx, &t_user.User{
    Name: "John",
})
```

### Update

```go
err := t_user.ORM.UpdateWhere(ctx, t_user.ID.Eq(123), &t_user.UserOptional{
    Name: ptr("NewName"),
})
```

### Delete

```go
err := t_user.ORM.DeleteWhere(ctx, t_user.ID.Eq(123))
```

## Route Handlers

Use `processor.Gin` for type-safe handlers with automatic request parsing:

```go
import "github.com/xhd2015/kool/tools/create/server_go_db_template/route/processor"

type CreateRequest struct {
    Name string `json:"name"`
}

type CreateResponse struct {
    ID int64 `json:"id"`
}

func Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
    // Business logic here
    return &CreateResponse{ID: 1}, nil
}

// Register: r.POST("/api/create", processor.Gin(Create))
```

The processor automatically handles:
- Request parsing (query params, JSON body, path params)
- Response formatting (wraps result in `{"code":0,"data":...}`)
- Error handling (returns `{"code":-1,"msg":"error message"}`)
- Session injection

## Error Handling

Return errors from handlers - they will be automatically converted to HTTP responses:

```go
func MyHandler(ctx context.Context, req *MyRequest) (*MyResponse, error) {
    if req.Name == "" {
        return nil, fmt.Errorf("name is required")
    }
    // Use predefined errors
    return nil, server_errors.ErrUserNotFound
}
```

## Unit Testing

### Testing Handlers

Use `httptest` with gin's test mode:

```go
package example_test

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
)

func TestGet(t *testing.T) {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.GET("/api/example", processor.Gin(example.Get))

    req := httptest.NewRequest("GET", "/api/example?id=1", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", w.Code)
    }

    var resp map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &resp)
    // Assert response...
}
```

### Testing Services

Test service functions directly with mocked dependencies:

```go
package user_test

import (
    "context"
    "testing"

    "github.com/xhd2015/kool/tools/create/server_go_db_template/service/user"
)

func TestGetUserByID(t *testing.T) {
    ctx := context.Background()
    // Setup test database or mock...

    result, err := user.GetUserByID(ctx, 123)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // Assert result...
}
```

### Test File Naming

- Place tests in `*_test.go` files in the same package or `<package>_test` package
- Use table-driven tests for multiple test cases
- Use `t.Helper()` in helper functions

## Session Management

```go
import "github.com/xhd2015/kool/tools/create/server_go_db_template/service/session"

func MyHandler(ctx context.Context, req *MyRequest) (*MyResponse, error) {
    sess := session.GetSession(ctx)
    userID := sess.GetUserID()
    // ...
}
```

## Naming Conventions

1. **Go identifiers** - Use camelCase
2. **Database columns** - Use snake_case
3. **API endpoints** - Use kebab-case
4. **Table packages** - Use `t_<table_name>` naming
5. **Error variables** - Prefix with `Err`

## Error Messages

- Start with lowercase
- Be descriptive but concise
- Include relevant context (IDs, values)

```go
// Good
return fmt.Errorf("user %d not found", userID)
return fmt.Errorf("failed to create order: %w", err)

// Bad
return fmt.Errorf("Error: User Not Found")
```
