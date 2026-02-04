package task

import (
	"context"
	"log"

	cron "github.com/robfig/cron/v3"
)

// Cron expressions
const (
	// EVERY_MINUTE runs every minute
	EVERY_MINUTE = "* * * * *"
	// EVERY_HOUR runs at minute 0 of every hour
	EVERY_HOUR = "0 * * * *"
	// EVERY_DAY_BEGIN runs at 00:00 every day
	EVERY_DAY_BEGIN = "0 0 * * *"
)

var cronTask *cron.Cron

// Init initializes the task scheduler
func Init(ctx context.Context) {
	cronTask = cron.New()
	RegisterCronTasks()
	cronTask.Start()
}

// addTask registers a cron task with the given expression
// See https://crontab.cronhub.io/ for cron expression syntax
//
// Examples:
//   - "* * * * *"     - every minute
//   - "0 * * * *"     - every hour at minute 0
//   - "0 0 * * *"     - every day at 00:00
//   - "0 */2 * * *"   - every 2 hours
//   - "0 0 * * 0"     - every Sunday at 00:00
func addTask(expr string, f func(ctx context.Context) error) {
	cronTask.AddFunc(expr, func() {
		err := f(context.Background())
		if err != nil {
			log.Printf("ERROR task: %v", err)
		}
	})
}

// Stop stops the cron scheduler
func Stop() {
	if cronTask != nil {
		cronTask.Stop()
	}
}
