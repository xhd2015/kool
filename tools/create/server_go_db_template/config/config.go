package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/xhd2015/kool/tools/create/server_go_db_template/env"
)

// Config represents the application configuration
type Config struct {
	// MySQL configuration
	MySQL MySQLConfig `json:"mysql"`

	// Log configuration
	Log LogConfig `json:"log"`

	// Server configuration
	Server ServerConfig `json:"server"`
}

// MySQLConfig represents MySQL database configuration
type MySQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// LogConfig represents logging configuration
type LogConfig struct {
	Path       string `json:"path"`
	Level      string `json:"level"`
	MaxSize    int    `json:"maxSize"`    // max size in megabytes before rotation
	MaxBackups int    `json:"maxBackups"` // max number of old log files to keep
	MaxAge     int    `json:"maxAge"`     // max days to retain old log files
	Compress   bool   `json:"compress"`   // whether to compress rotated files
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port      int    `json:"port"`
	StaticDir string `json:"staticDir"`
	AllowCORS bool   `json:"allowCors"`
}

// global config instance
var globalConfig *Config

// Load loads configuration from a JSON file
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Get returns the global config instance
// Returns nil if config is not loaded
func Get() *Config {
	return globalConfig
}

// Set sets the global config instance
func Set(cfg *Config) {
	globalConfig = cfg
}

// GetMySQL returns MySQL config, with defaults if not set
func (c *Config) GetMySQL() MySQLConfig {
	cfg := c.MySQL
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == 0 {
		cfg.Port = 3306
	}
	if cfg.User == "" {
		cfg.User = "root"
	}
	if cfg.Database == "" {
		cfg.Database = "template_db"
	}
	return cfg
}

// GetServer returns Server config, with defaults if not set
func (c *Config) GetServer() ServerConfig {
	cfg := c.Server
	if cfg.Port == 0 {
		cfg.Port = 8008
	}
	return cfg
}

// GetLog returns Log config, with defaults if not set
func (c *Config) GetLog() LogConfig {
	cfg := c.Log
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.MaxSize == 0 {
		cfg.MaxSize = 100 // 100MB default
	}
	if cfg.MaxBackups == 0 {
		cfg.MaxBackups = 3
	}
	if cfg.MaxAge == 0 {
		cfg.MaxAge = 28 // 28 days default
	}
	return cfg
}

// Flags holds command line flag values for building effective config
type Flags struct {
	Port          int
	Static        string
	AllowCORS     bool
	MySQLHost     string
	MySQLPort     int
	MySQLUser     string
	MySQLPassword string
	MySQLDatabase string
	LogPath       string
	LogLevel      string
	LogMaxSize    int
	LogMaxBackups int
	LogMaxAge     int
	LogCompress   bool
}

// BuildEffectiveConfig builds the effective configuration with priority:
// command line flags > config file > environment variables > defaults
func BuildEffectiveConfig(fileCfg *Config, cmdFlags Flags) *Config {
	cfg := &Config{}

	// Server config
	cfg.Server.Port = resolveInt(cmdFlags.Port, getConfigInt(fileCfg, func(c *Config) int { return c.Server.Port }), env.GetMySQLPort(), 8008)
	cfg.Server.StaticDir = resolveString(cmdFlags.Static, getConfigString(fileCfg, func(c *Config) string { return c.Server.StaticDir }), "", "")
	cfg.Server.AllowCORS = resolveBool(cmdFlags.AllowCORS, getConfigBool(fileCfg, func(c *Config) bool { return c.Server.AllowCORS }), env.AllowCORS())

	// MySQL config
	cfg.MySQL.Host = resolveString(cmdFlags.MySQLHost, getConfigString(fileCfg, func(c *Config) string { return c.MySQL.Host }), "", "localhost")
	cfg.MySQL.Port = resolveInt(cmdFlags.MySQLPort, getConfigInt(fileCfg, func(c *Config) int { return c.MySQL.Port }), env.GetMySQLPort(), 3306)
	cfg.MySQL.User = resolveString(cmdFlags.MySQLUser, getConfigString(fileCfg, func(c *Config) string { return c.MySQL.User }), "", "root")
	cfg.MySQL.Password = resolveString(cmdFlags.MySQLPassword, getConfigString(fileCfg, func(c *Config) string { return c.MySQL.Password }), "", "")
	cfg.MySQL.Database = resolveString(cmdFlags.MySQLDatabase, getConfigString(fileCfg, func(c *Config) string { return c.MySQL.Database }), "", "template_db")

	// Log config
	cfg.Log.Path = resolveString(cmdFlags.LogPath, getConfigString(fileCfg, func(c *Config) string { return c.Log.Path }), "", "")
	cfg.Log.Level = resolveString(cmdFlags.LogLevel, getConfigString(fileCfg, func(c *Config) string { return c.Log.Level }), "", "info")
	cfg.Log.MaxSize = resolveInt(cmdFlags.LogMaxSize, getConfigInt(fileCfg, func(c *Config) int { return c.Log.MaxSize }), 0, 100)
	cfg.Log.MaxBackups = resolveInt(cmdFlags.LogMaxBackups, getConfigInt(fileCfg, func(c *Config) int { return c.Log.MaxBackups }), 0, 3)
	cfg.Log.MaxAge = resolveInt(cmdFlags.LogMaxAge, getConfigInt(fileCfg, func(c *Config) int { return c.Log.MaxAge }), 0, 28)
	cfg.Log.Compress = resolveBool(cmdFlags.LogCompress, getConfigBool(fileCfg, func(c *Config) bool { return c.Log.Compress }), false)

	return cfg
}

func getConfigInt(cfg *Config, getter func(*Config) int) int {
	if cfg == nil {
		return 0
	}
	return getter(cfg)
}

func getConfigString(cfg *Config, getter func(*Config) string) string {
	if cfg == nil {
		return ""
	}
	return getter(cfg)
}

func getConfigBool(cfg *Config, getter func(*Config) bool) bool {
	if cfg == nil {
		return false
	}
	return getter(cfg)
}

// resolveInt returns the first non-zero value in priority order
func resolveInt(cmdLine, configFile, envVar, defaultVal int) int {
	if cmdLine != 0 {
		return cmdLine
	}
	if configFile != 0 {
		return configFile
	}
	if envVar != 0 {
		return envVar
	}
	return defaultVal
}

// resolveString returns the first non-empty value in priority order
func resolveString(cmdLine, configFile, envVar, defaultVal string) string {
	if cmdLine != "" {
		return cmdLine
	}
	if configFile != "" {
		return configFile
	}
	if envVar != "" {
		return envVar
	}
	return defaultVal
}

// resolveBool returns true if any source is true (command line has highest priority)
func resolveBool(cmdLine, configFile, envVar bool) bool {
	if cmdLine {
		return true
	}
	if configFile {
		return true
	}
	return envVar
}
