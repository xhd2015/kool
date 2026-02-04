package task

import (
	"context"

	"github.com/xhd2015/kool/tools/create/server_go_db_template/lib/log"
)

// ExampleCleanupTask demonstrates how to create a cron task
// This task runs every hour and performs cleanup operations
//
// Usage:
//
//	In cron_tasks.go:
//	addTask(EVERY_HOUR, ExampleCleanupTask)
func ExampleCleanupTask(ctx context.Context) error {
	log.Infof(ctx, "running example cleanup task")

	// Add your task logic here
	// For example:
	// - Clean up expired sessions
	// - Archive old records
	// - Send scheduled notifications
	// - Sync data with external services

	log.Infof(ctx, "example cleanup task completed")
	return nil
}
