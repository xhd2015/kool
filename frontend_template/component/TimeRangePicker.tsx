import { DatePicker } from "antd"
import type { Dayjs } from 'dayjs'
import dayjs from 'dayjs'
import { use, useEffect, useState } from "react"
import { useSearchParams } from 'react-router'
import { useCurrent } from "../util/useCurrent"

const { RangePicker } = DatePicker

export interface RelativeTime {
    num: number
    unit: 'minute' | 'hour' | 'day' | 'week' | 'month'
}

export interface TimeRangePickerProps {
    bindSearchParams?: boolean
    defaultRelativeTime?: RelativeTime
    value?: [Date, Date]
    onChange: (value: [Date, Date] | null) => void
}

const presets: {
    label: string
    value: [Dayjs, Dayjs]
}[] = [
        { label: 'Last 5 minutes', value: [dayjs().subtract(5, 'minute'), dayjs()] },
        { label: 'Last 1 hour', value: [dayjs().subtract(1, 'hour'), dayjs()] },
        { label: 'Last 2 hours', value: [dayjs().subtract(2, 'hour'), dayjs()] },
        { label: 'Last 6 hours', value: [dayjs().subtract(3, 'hour'), dayjs()] },
        { label: 'Last 12 hours', value: [dayjs().subtract(12, 'hour'), dayjs()] },
        { label: 'Last 24 hours', value: [dayjs().subtract(24, 'hour'), dayjs()] },
        { label: 'Last 1 week', value: [dayjs().subtract(1, 'week'), dayjs()] },
        { label: 'Last 2 weeks', value: [dayjs().subtract(2, 'week'), dayjs()] },
        { label: 'Last 1 month', value: [dayjs().subtract(1, 'month'), dayjs()] },
    ]

// data flow: 
// if bindSearchParams:  value is effectively ignored, the time range is determined by the URL parameters
//    - url changed: url param -> TimeRangePicker.timeRange -> RangePicker(UI) + onChange(external)
//    - user changed time range: user -> RangePicker(UI) + handleTimeRangeChange -> url param
//
// if !bindSearchParams: value is used, and the time range is determined by the value

export function TimeRangePicker({ value, bindSearchParams, defaultRelativeTime, onChange }: TimeRangePickerProps) {
    const [searchParams, setSearchParams] = useSearchParams()

    const [timeRange, setTimeRange] = useState<[Dayjs, Dayjs] | null>(() => {
        if (bindSearchParams) {
            // Try to get start and end from URL parameters
            const start = searchParams.get('start')
            const end = searchParams.get('end')
            if (start && end) {
                const startDayjs = dayjs(start)
                const endDayjs = dayjs(end)
                if (startDayjs.isValid() && endDayjs.isValid()) {
                    return [startDayjs, endDayjs]
                }
            }
        } else if (value) {
            return [dayjs(value[0]), dayjs(value[1])]
        }
        // Default to last 24 hours if no valid params
        const endTime = dayjs()
        let relTime = defaultRelativeTime || { num: 24, unit: 'hour' }
        const startTime = endTime.subtract(relTime.num, relTime.unit)
        return [startTime, endTime]
    })

    const startTime = timeRange?.[0]?.toDate?.()?.getTime?.()
    const endTime = timeRange?.[1]?.toDate?.()?.getTime?.()

    const onChangeRef = useCurrent(onChange)
    const timeRangeRef = useCurrent(timeRange)

    // Add effect to listen for URL parameter changes
    useEffect(() => {
        if (bindSearchParams) {
            const start = searchParams.get('start')
            const end = searchParams.get('end')
            if (start && end) {
                const startDayjs = dayjs(start)
                const endDayjs = dayjs(end)
                if (startDayjs.isValid() && endDayjs.isValid() &&
                    (!timeRangeRef.current ||
                        startDayjs.format() !== timeRangeRef.current[0].format() ||
                        endDayjs.format() !== timeRangeRef.current[1].format())) {
                    setTimeRange([startDayjs, endDayjs])
                }
            }
            return
        }
        if (value) {
            setTimeRange([dayjs(value[0]), dayjs(value[1])])
        }
    }, [searchParams, value])

    // Update URL when time range changes manually
    const handleTimeRangeChange = (range: [Dayjs, Dayjs] | null) => {
        if (!bindSearchParams) {
            setTimeRange(range)
            return
        }
        if (range) {
            const [start, end] = range
            setSearchParams({
                ...Object.fromEntries(searchParams.entries()),
                start: start.format(),
                end: end.format(),
            }, {
                replace: true
            })
        } else {
            // delete start and end from url params
            const newParams = new URLSearchParams(searchParams)
            newParams.delete('start')
            newParams.delete('end')
            setSearchParams(newParams, {
                replace: true
            })
        }
    }

    useEffect(() => {
        if (onChangeRef.current && timeRangeRef.current) {
            const start = timeRangeRef.current[0].toDate()
            const end = timeRangeRef.current[1].toDate()
            // console.log("time range changed", start, end)
            onChangeRef.current([start, end])
        }
    }, [startTime, endTime])

    return <RangePicker
        showTime
        preserveInvalidOnBlur
        needConfirm={false}
        value={timeRange}
        onChange={handleTimeRangeChange}
        presets={presets.map(preset => ({
            label: preset.label,
            value: preset.value
        }))}
    />
}
