import dayjs from 'dayjs'

export const HOUR_MS = 1 * 60 * 60 * 1000

export function formatServerDatetime(datetime: Date) {
    // format with correct timezone
    return dayjs(datetime).format('YYYY-MM-DDTHH:mm:ssZ')
}

export function formatDate(datetime: Date) {
    // format with correct timezone
    return dayjs(datetime).format('YYYY-MM-DD')
}
export function formatDateNoep(datetime: Date) {
    // format with correct timezone
    return dayjs(datetime).format('YYYYMMDD')
}

export interface TimeRange {
    start: Date
    end: Date
}

export function splitTimeRange(start: Date, end: Date, stepMs: number): TimeRange[] {
    const ranges: TimeRange[] = []

    let startMs = start.getTime()
    let endMs = end.getTime()

    let curMs = startMs
    while (curMs < endMs) {
        let nextMs = curMs + stepMs
        if (nextMs > endMs) {
            nextMs = endMs
        }
        ranges.push({ start: new Date(curMs), end: new Date(nextMs) })
        curMs = nextMs
    }
    return ranges
}