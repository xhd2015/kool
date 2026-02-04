// Package env provides centralized access to environment variables used by the server.
package env

import (
	"os"
	"strconv"
)

// Environment variable names
const (
	// Server configuration
	EnvAllowCORS = "SERVER_ALLOW_CORS"

	// Database configuration
	EnvMySQLHost     = "MYSQL_HOST"
	EnvMySQLPort     = "MYSQL_PORT"
	EnvMySQLUser     = "MYSQL_USER"
	EnvMySQLPassword = "MYSQL_PASSWORD"
	EnvMySQLDatabase = "MYSQL_DATABASE"
)

// AllowCORS returns true if CORS is enabled via environment variable.
func AllowCORS() bool {
	return os.Getenv(EnvAllowCORS) == "true"
}

// MySQLHost returns the MySQL host from environment, or empty string if not set.
func MySQLHost() string {
	return os.Getenv(EnvMySQLHost)
}

// MySQLPort returns the MySQL port from environment, or empty string if not set.
func MySQLPort() string {
	return os.Getenv(EnvMySQLPort)
}

// GetMySQLPort returns the MySQL port as int from environment, or 0 if not set.
func GetMySQLPort() int {
	portStr := os.Getenv(EnvMySQLPort)
	if portStr == "" {
		return 0
	}
	port, _ := strconv.Atoi(portStr)
	return port
}

// MySQLUser returns the MySQL user from environment, or empty string if not set.
func MySQLUser() string {
	return os.Getenv(EnvMySQLUser)
}

// MySQLPassword returns the MySQL password from environment, or empty string if not set.
func MySQLPassword() string {
	return os.Getenv(EnvMySQLPassword)
}

// MySQLDatabase returns the MySQL database name from environment, or empty string if not set.
func MySQLDatabase() string {
	return os.Getenv(EnvMySQLDatabase)
}
