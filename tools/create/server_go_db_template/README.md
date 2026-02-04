# Go Server Template

A minimal Go project template with Gin server and xorm/arc-orm database integration.

## Features

- **Gin HTTP Server** - Fast and lightweight web framework
- **xorm/arc-orm** - Database ORM integration with MySQL
- **Session Management** - Basic session handling
- **Structured Logging** - Context-aware logging with slog + zap backend
- **Prometheus Metrics** - Built-in metrics endpoint
- **Podman/Docker Support** - MySQL container setup included

## Quick Start

### Prerequisites

- Go 1.22+
- Podman or Docker (for MySQL)

### Start Development Server

```bash
# Start MySQL container
go run ./script/podman-compose

# Wait for MySQL to be ready
go run ./script/mysql check

# Start the server
go run ./script/dev
```

Or use VSCode tasks: `Cmd+Shift+B` → "Start Server (with DB)"

### Access Local MySQL

After starting the MySQL container, you can access the MySQL client:

```bash
go run ./script/mysql
```

This will open an interactive MySQL shell connected to the local database.

### Build

```bash
go run ./script/build
```

## Project Structure

```
.
├── main.go              # Application entry point
├── config/              # Configuration management
├── route/               # Route definitions
├── dao/                 # Database access layer
├── service/             # Business logic
├── lib/                 # Shared utilities
│   ├── log/             # Structured logging (slog + zap)
│   ├── metrics/         # Prometheus metrics
│   ├── middleware/      # HTTP middlewares
│   ├── routehelp/       # Route helpers
│   └── routewrap/       # Gin wrappers
├── env/                 # Environment configuration
├── types/               # Type definitions
└── script/              # Build and utility scripts
    ├── dev/             # Development server
    ├── build/           # Build binary
    ├── mysql/           # MySQL client access
    ├── podman-compose/  # Container management
    └── podman-machine-start/  # Podman machine startup
```

## Configuration

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `MYSQL_HOST` | MySQL host | localhost |
| `MYSQL_PORT` | MySQL port | 3306 |
| `MYSQL_USER` | MySQL user | app |
| `MYSQL_PASSWORD` | MySQL password | apppassword |
| `MYSQL_DATABASE` | MySQL database | app_db |
| `ALLOW_CORS` | Enable CORS | false |

## License

MIT
