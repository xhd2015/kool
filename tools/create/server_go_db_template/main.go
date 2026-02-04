package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/config"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/dao"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/lib/log"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/lib/metrics"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/lib/middleware"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/route"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/task"
	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage:
  server [flags]

Flags:
  --config string         Path to config file
  --port int              Port to listen on (default: 8008)
  --static string         Static file directory
  --allow-cors            Allow CORS requests (default: false)
  --mysql-host string     MySQL host (default: localhost)
  --mysql-port int        MySQL port (default: 3306)
  --mysql-user string     MySQL user (default: root)
  --mysql-password string MySQL password
  --mysql-database string MySQL database (default: template_db)
  --log-path string       Log file path
  --log-level string      Log level (default: info)
  --log-max-size int      Max log file size in MB (default: 100)
  --log-max-backups int   Max number of old log files (default: 3)
  --log-max-age int       Max days to retain old log files (default: 28)
  --log-compress          Compress rotated log files
  -h, --help              Show this help message
`

func Init(ctx context.Context) {
	dao.Init()
	task.Init(ctx)
}

func main() {
	args := os.Args[1:]
	err := start(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func start(args []string) error {
	// Command line flags
	var configPath string
	var cmdFlags config.Flags

	args, err := flags.String("--config", &configPath).
		// Server flags
		Int("--port", &cmdFlags.Port).
		String("--static", &cmdFlags.Static).
		Bool("--allow-cors", &cmdFlags.AllowCORS).
		// MySQL flags
		String("--mysql-host", &cmdFlags.MySQLHost).
		Int("--mysql-port", &cmdFlags.MySQLPort).
		String("--mysql-user", &cmdFlags.MySQLUser).
		String("--mysql-password", &cmdFlags.MySQLPassword).
		String("--mysql-database", &cmdFlags.MySQLDatabase).
		// Log flags
		String("--log-path", &cmdFlags.LogPath).
		String("--log-level", &cmdFlags.LogLevel).
		Int("--log-max-size", &cmdFlags.LogMaxSize).
		Int("--log-max-backups", &cmdFlags.LogMaxBackups).
		Int("--log-max-age", &cmdFlags.LogMaxAge).
		Bool("--log-compress", &cmdFlags.LogCompress).
		// Help
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return fmt.Errorf("unexpected extra args: %v", args)
	}

	// Load config file if specified
	var cfg *config.Config
	if configPath != "" {
		cfg, err = config.Load(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	// Build effective config with priority: command line > config file > env
	effectiveCfg := config.BuildEffectiveConfig(cfg, cmdFlags)

	return Run(effectiveCfg)
}

// Run starts the server with the given configuration
func Run(cfg *config.Config) error {
	// Set global config
	config.Set(cfg)

	// Initialize log with rotation
	logCfg := cfg.GetLog()
	log.Init(log.Config{
		Path:       logCfg.Path,
		Level:      logCfg.Level,
		MaxSize:    logCfg.MaxSize,
		MaxBackups: logCfg.MaxBackups,
		MaxAge:     logCfg.MaxAge,
		Compress:   logCfg.Compress,
	})

	ctx := context.Background()
	Init(ctx)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Trace middleware - inject trace ID into context
	r.Use(middleware.Trace())

	// Metrics middleware - record HTTP metrics
	r.Use(metrics.Middleware())

	// Recovery middleware
	r.Use(middleware.Recovery())

	// CORS middleware
	if cfg.Server.AllowCORS {
		r.Use(middleware.CORS())
	}

	// Static files middleware
	if cfg.Server.StaticDir != "" {
		r.Use(middleware.StaticFiles(cfg.Server.StaticDir))
	}

	// Auth middleware
	r.Use(middleware.Auth(middleware.AuthConfig{
		PublicPaths: []string{"/api/login", "/api/register", "/ping", "/metrics"},
	}))

	// Health check
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, map[string]interface{}{
			"data": "pong",
		})
	})

	// Metrics endpoint
	r.GET("/metrics", gin.WrapH(metrics.Handler()))

	// Initialize routes
	route.Init(r)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Fprintf(os.Stderr, "Server listen at http://localhost:%d\n", cfg.Server.Port)
	return r.Run(addr)
}
