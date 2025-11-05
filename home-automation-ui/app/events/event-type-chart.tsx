"use client"

import { useMemo, useId } from "react"
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart"
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Event, EventType } from "@/types/events"
import { capitalize } from "@/lib/utils"

type TimeInterval = "1h" | "6h" | "12h" | "24h" | "7d"

const TIME_INTERVALS: { value: TimeInterval; label: string; hours: number }[] = [
  { value: "1h", label: "Last Hour", hours: 1 },
  { value: "6h", label: "Last 6 Hours", hours: 6 },
  { value: "12h", label: "Last 12 Hours", hours: 12 },
  { value: "24h", label: "Last 24 Hours", hours: 24 },
  { value: "7d", label: "Last 7 Days", hours: 168 },
]

interface EventTypeChartProps {
  events: Event[]
  hours: number
  onHoursChange: (hours: number) => void
}

interface TimeBucket {
  time: string
  timestamp: number
  [key: string]: string | number
}

export function EventTypeChart({ events, hours, onHoursChange }: EventTypeChartProps) {
  const chartId = useId().replace(/:/g, "")
  const selectedInterval = TIME_INTERVALS.find((i) => i.hours === hours) || TIME_INTERVALS[3]
  const { chartData, chartConfig, eventTypes } = useMemo(() => {
    // Show selected time interval
    const now = new Date()
    const intervalStart = new Date(now.getTime() - hours * 60 * 60 * 1000)

    // Get unique event types
    const uniqueTypes = Array.from(new Set<EventType>(events.map((e) => e.type)))
    
    if (uniqueTypes.length === 0) {
      return { chartData: [], chartConfig: {}, eventTypes: [] }
    }

    // Determine bucket size based on interval
    let bucketMs: number
    if (hours <= 1) {
      bucketMs = 5 * 60 * 1000 // 5-minute buckets for 1 hour
    } else if (hours <= 6) {
      bucketMs = 15 * 60 * 1000 // 15-minute buckets for 6 hours
    } else if (hours <= 24) {
      bucketMs = 30 * 60 * 1000 // 30-minute buckets for 24 hours
    } else {
      bucketMs = 4 * 60 * 60 * 1000 // 4-hour buckets for 7 days
    }

    // Create all time buckets for the selected interval
    const buckets = new Map<number, Record<EventType, number>>()
    
    // Initialize all buckets for the selected interval
    const startTime = Math.floor(intervalStart.getTime() / bucketMs) * bucketMs
    const endTime = Math.floor(now.getTime() / bucketMs) * bucketMs
    
    for (let bucketTime = startTime; bucketTime <= endTime; bucketTime += bucketMs) {
      const initialCounts = {} as Record<EventType, number>
      uniqueTypes.forEach((type) => {
        initialCounts[type] = 0
      })
      buckets.set(bucketTime, initialCounts)
    }

    // Populate buckets with actual event data
    events.forEach((event) => {
      const eventTime = new Date(event.time_fired).getTime()
      const bucketTime = Math.floor(eventTime / bucketMs) * bucketMs
      
      if (buckets.has(bucketTime)) {
        const bucket = buckets.get(bucketTime)!
        bucket[event.type] = (bucket[event.type] || 0) + 1
      }
    })

    // Convert to array and format for chart
    const data: TimeBucket[] = Array.from(buckets.entries())
      .sort(([a], [b]) => a - b)
      .map(([bucketTime, counts]) => {
        const time = new Date(bucketTime)
        const formattedTime = hours === 168
          ? time.toLocaleDateString("en-US", {
              month: "short",
              day: "numeric",
            })
          : time.toLocaleTimeString("en-GB", {
              hour: "2-digit",
              minute: "2-digit",
            })

        const bucket: TimeBucket = { time: formattedTime, timestamp: bucketTime }

        uniqueTypes.forEach((type) => {
          const displayName = type
            .split("_")
            .map((word) => capitalize(word))
            .join(" ")
          bucket[displayName] = counts[type] || 0
        })

        return bucket
      })

    // Create chart config for each event type
    const config: Record<string, { label: string; color: string }> = {}
    uniqueTypes.forEach((type) => {
      const displayName = type
        .split("_")
        .map((word) => capitalize(word))
        .join(" ")
      config[displayName] = {
        label: displayName,
        color: getColorForType(type),
      }
    })

    return {
      chartData: data,
      chartConfig: config,
      eventTypes: uniqueTypes.map((type) => ({
        type,
        displayName: type
          .split("_")
          .map((word) => capitalize(word))
          .join(" "),
      })),
    }
  }, [events, hours])

  if (chartData.length === 0 || eventTypes.length === 0) {
    return null
  }

  // Create labelFormatter for tooltip that prepends date if not today
  const tooltipLabelFormatter = (value: string | undefined, payload: readonly any[]) => {
    if (!payload || payload.length === 0) return value || ""
    
    const dataPayload = payload[0]?.payload as TimeBucket | undefined
    if (!dataPayload?.timestamp) return value || ""
    
    const dataDate = new Date(dataPayload.timestamp)
    const today = new Date()
    const isToday = 
      dataDate.getDate() === today.getDate() &&
      dataDate.getMonth() === today.getMonth() &&
      dataDate.getFullYear() === today.getFullYear()
    
    if (isToday || hours === 168) {
      // For today or 7-day view (which already shows dates), just return the value
      return value || ""
    }
    
    // For other dates, prepend the date
    const dateStr = dataDate.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
    })
    return `${dateStr} ${value || ""}`
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <CardTitle>Event Timeline</CardTitle>
            <CardDescription>
              Number of events received over the selected time interval by type
            </CardDescription>
          </div>
          <Select
            value={selectedInterval.value}
            onValueChange={(value) => {
              const interval = TIME_INTERVALS.find((i) => i.value === value as TimeInterval)
              if (interval) {
                onHoursChange(interval.hours)
              }
            }}
          >
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Select time interval" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectLabel>Time Interval</SelectLabel>
                {TIME_INTERVALS.map((interval) => (
                  <SelectItem key={interval.value} value={interval.value}>
                    {interval.label}
                  </SelectItem>
                ))}
              </SelectGroup>
            </SelectContent>
          </Select>
        </div>
      </CardHeader>
      <CardContent>
        <div>
          <ChartContainer config={chartConfig} className="h-[300px] w-full">
            <AreaChart data={chartData} margin={{ right: 30 }}>
              <defs>
                {eventTypes.map(({ type }) => {
                  const color = getColorForType(type)
                  const gradientId = `gradient-${chartId}-${type}`
                  return (
                    <linearGradient key={gradientId} id={gradientId} x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor={color} stopOpacity={0.8} />
                      <stop offset="95%" stopColor={color} stopOpacity={0.1} />
                    </linearGradient>
                  )
                })}
              </defs>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis
                dataKey="time"
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                tickFormatter={(value) => value}
                interval={chartData.length <= 7 ? 0 : Math.floor((chartData.length - 1) / 7)}
              />
              <YAxis
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                allowDecimals={false}
                domain={[0, (dataMax) => Math.max(1, dataMax * 1.1)]}
              />
              <ChartTooltip
                cursor={{ strokeDasharray: "3 3" }}
                content={<ChartTooltipContent indicator="line" labelFormatter={tooltipLabelFormatter} />}
              />
              {eventTypes.map(({ displayName, type }) => {
                const gradientId = `gradient-${chartId}-${type}`
                return (
                  <Area
                    key={type}
                    type="linear"
                    dataKey={displayName}
                    stroke={getColorForType(type)}
                    strokeWidth={2}
                    fill={`url(#${gradientId})`}
                    stackId="1"
                    baseValue={0}
                  />
                )
              })}
            </AreaChart>
          </ChartContainer>
          <div className="flex flex-wrap justify-center gap-4 text-xs mt-4">
            {eventTypes.map(({ displayName, type }) => (
              <div
                key={type}
                className="flex items-center gap-1.5 [&_svg]:pointer-events-none [&_svg]:size-3 [&_svg]:shrink-0"
              >
                <div
                  className="size-2 shrink-0 rounded-[2px]"
                  style={{
                    backgroundColor: getColorForType(type),
                  }}
                />
                <span className="text-muted-foreground">{displayName}</span>
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

function getColorForType(type: EventType): string {
  switch (type) {
    case "state_changed":
      return "hsl(217, 91%, 60%)" // Darker blue
    case "call_service":
      return "hsl(217, 98%, 78%)" // Lighter blue
    case "time_changed":
      return "hsl(217, 98%, 78%)" // Light blue
    default:
      return "hsl(217, 91%, 69%)" // Medium blue
  }
}

