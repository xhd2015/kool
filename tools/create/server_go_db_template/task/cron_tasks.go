package task

// RegisterCronTasks registers all cron tasks
// Add your tasks here using addTask()
//
// Example:
//
//	addTask(EVERY_HOUR, MyHourlyTask)
//	addTask(EVERY_DAY_BEGIN, MyDailyTask)
func RegisterCronTasks() {
	// Example: run cleanup task every hour
	addTask(EVERY_HOUR, ExampleCleanupTask)
}
