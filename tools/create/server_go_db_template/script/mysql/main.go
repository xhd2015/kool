package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/xgo/support/cmd"
	"golang.org/x/term"
)

const help = `
Usage:
  mysql [command] [flags]

Commands:
  (default)  Enter MySQL client in container
  check      Check if MySQL is ready (waits up to 2 minutes)

Flags:
  --host string       MySQL host (default: localhost, env: MYSQL_HOST)
  --port string       MySQL port (default: 3306, env: MYSQL_PORT)
  --user string       MySQL user (default: app, env: MYSQL_USER)
  --password string   MySQL password (default: apppassword, env: MYSQL_PASSWORD)
  --database string   MySQL database (default: app_db, env: MYSQL_DATABASE)
  --container string  Container name (default: app_db)
  --timeout duration  Connection timeout for check (default: 2m)
  -v, --verbose       Verbose output
  -h, --help          Show help

Example:
  go run ./script/mysql
  go run ./script/mysql check
  go run ./script/mysql check --host 127.0.0.1 --port 3306
`

func main() {
	err := Handle(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func Handle(args []string) error {
	var command string
	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}

	switch strings.ReplaceAll(command, "-", "_") {
	case "", "local":
		return handleLocal(args)
	case "check":
		return handleCheck(args)
	case "help", "__help", "--help":
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return nil
	default:
		return fmt.Errorf("unrecognized command: %s, use --help for usage", command)
	}
}

func handleLocal(args []string) error {
	var container string

	args, err := flags.
		Help("-h,--help", help).
		String("--container", &container).
		Parse(args)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return fmt.Errorf("unexpected extra args: %v", args)
	}

	if container == "" {
		container = "app_db"
	}

	isTTY := term.IsTerminal(int(os.Stdin.Fd()))
	execFlags := "-i"
	if isTTY {
		execFlags = "-it"
	}

	// podman exec -it app_db bash -c "mysql -uroot -prootpassword app_db"
	return cmd.Debug().Stdin(os.Stdin).Run("podman", "exec", execFlags, container, "bash", "-c", "mysql -uroot -prootpassword app_db")
}

func handleCheck(args []string) error {
	var host, port, user, password, database string
	var timeout time.Duration
	var verbose bool

	args, err := flags.
		Help("-h,--help", help).
		String("--host", &host).
		String("--port", &port).
		String("--user", &user).
		String("--password", &password).
		String("--database", &database).
		Duration("--timeout", &timeout).
		Bool("-v,--verbose", &verbose).
		Parse(args)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return fmt.Errorf("unexpected extra args: %v", args)
	}

	// Apply defaults from environment or hardcoded values
	if host == "" {
		host = getEnvOrDefault("MYSQL_HOST", "localhost")
	}
	if port == "" {
		port = getEnvOrDefault("MYSQL_PORT", "3306")
	}
	if user == "" {
		user = getEnvOrDefault("MYSQL_USER", "app")
	}
	if password == "" {
		password = getEnvOrDefault("MYSQL_PASSWORD", "apppassword")
	}
	if database == "" {
		database = getEnvOrDefault("MYSQL_DATABASE", "app_db")
	}
	if timeout == 0 {
		timeout = 2 * time.Minute
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", user, password, host, port, database)
	if verbose {
		fmt.Printf("Connecting to MySQL: %s:%s/%s as %s\n", host, port, database, user)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	begin := time.Now()
	for {
		pingErr := db.Ping()
		if pingErr == nil {
			fmt.Println("MySQL is ready!")
			return nil
		}
		if verbose {
			fmt.Printf("Waiting for MySQL: %v\n", pingErr)
		}
		if time.Since(begin) > timeout {
			return fmt.Errorf("MySQL connection timeout after %v", timeout)
		}
		time.Sleep(5 * time.Second)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
