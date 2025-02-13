// Utility to create a pool of concurrent tasks
export async function throttled<T, R>(tasks: T[], maxConcurrent: number, fn: (task: T) => Promise<R>): Promise<R[]> {
    if (tasks.length === 0) return []

    const results: (R | undefined)[] = new Array(tasks.length)
    const executing = new Set<Promise<void>>()

    // Process tasks until all are done
    for (let i = 0; i < tasks.length; i++) {
        // If we've hit the concurrency limit, wait for one to finish
        if (executing.size >= maxConcurrent) {
            await Promise.race(executing)
        }

        // Start a new task
        const task = tasks[i]
        const p = (async () => {
            try {
                results[i] = await fn(task)
            } catch (err) {
                // Just log the error, the task function should handle errors
                console.error('Task failed:', err)
            }
        })()

        // Add to executing set and remove when done
        executing.add(p)
        p.finally(() => executing.delete(p))
    }

    // Wait for remaining tasks
    if (executing.size > 0) {
        await Promise.all(executing)
    }

    return results.filter((r): r is R => r !== undefined)
}
