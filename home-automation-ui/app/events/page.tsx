"use client"

import { useEffect, useMemo, useState } from "react"
import { createColumns } from "./columns"
import { DataTable } from "./data-table"
import { onSocketMessage } from "@/lib/engine-socket"
import { Event, EventType, DomainSpecificEventData } from "@/types/events"
import { Skeleton } from "@/components/ui/skeleton"
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from "@/components/ui/select"
import { SummaryStatCard } from "@/components/common/summary-stat-card"
import { useSwrLite } from "@/lib/swr-lite"
import { capitalize } from "@/lib/utils"
import { EventTypeChart } from "./event-type-chart"
import { Input } from "@/components/ui/input"
import { SearchIcon, XIcon } from "lucide-react"
import { Switch } from "@/components/ui/switch"
import { Label } from "@/components/ui/label"
import { isPending } from "./causality-chain-utils"


function parseEventMessage(msg: any): Event {
  const raw = typeof msg === "string" ? JSON.parse(msg) : msg

  // Extract context_id from context.id if not directly available
  const contextId = raw.context_id ?? raw.ContextID ?? raw.context?.id ?? raw.Context?.id ?? ""
  
  // Use context.id as fallback for id field
  const id = raw.id ?? raw.ID ?? raw.context?.id ?? raw.Context?.id ?? ""

  return {
    id: id,
    type: raw.type ?? raw.Type ?? "unknown",
    data: raw.data ?? raw.Data ?? {},
    context_id: contextId,
    time_fired: raw.time_fired ?? raw.TimeFired ?? new Date().toISOString(),
    context: raw.context ?? raw.Context ?? null,
  }
}

type TimeInterval = "1h" | "6h" | "12h" | "24h" | "7d"

const TIME_INTERVALS: { value: TimeInterval; label: string; hours: number }[] = [
  { value: "1h", label: "Last Hour", hours: 1 },
  { value: "6h", label: "Last 6 Hours", hours: 6 },
  { value: "12h", label: "Last 12 Hours", hours: 12 },
  { value: "24h", label: "Last 24 Hours", hours: 24 },
  { value: "7d", label: "Last 7 Days", hours: 168 },
]

export default function EventsPage() {
  const [typeFilter, setTypeFilter] = useState<string>("all")
  const [timeInterval, setTimeInterval] = useState<TimeInterval>("24h")
  const [searchQuery, setSearchQuery] = useState("")
  const [hideOptimistic, setHideOptimistic] = useState(false)

  const selectedInterval = TIME_INTERVALS.find((i) => i.value === timeInterval) || TIME_INTERVALS[3]

  const { data: eventsData, loading, mutate: mutateEvents, revalidate } = useSwrLite<Event[]>(
    `/api/events?hours=${selectedInterval.hours}`,
    async () => {
      const res = await fetch(`/api/events?hours=${selectedInterval.hours}`, { cache: "no-store" })
      if (!res.ok) {
        console.error("Failed to fetch events:", res.statusText)
        return []
      }
      const data = (await res.json()) as Event[]
      return data
    },
    { ttlMs: 5 * 60_000, revalidateOnFocus: true }
  )

  const events = eventsData ?? []

  // Check if there are any pending service calls using existing logic
  const hasPendingEvents = useMemo(() => {
    return events.some(event => isPending(event, events))
  }, [events])

  // Conditional polling: poll more frequently when there are pending events
  useEffect(() => {
    // When there are pending events, poll every 2 seconds
    // When there are no pending events, poll every 30 seconds (background refresh)
    const pollingInterval = hasPendingEvents ? 2000 : 30000
    
    const interval = setInterval(() => {
      revalidate()
    }, pollingInterval)
    
    return () => clearInterval(interval)
  }, [hasPendingEvents, revalidate])

  useEffect(() => {
    const handleMessage = (message: MessageEvent) => {
      try {
        const event = parseEventMessage(message.data)
        mutateEvents((prev?: Event[]) => {
          const updated = prev ? [event, ...prev] : [event]
          // Filter to selected time interval and limit to prevent memory issues
          const now = new Date()
          const intervalStart = new Date(now.getTime() - selectedInterval.hours * 60 * 60 * 1000)
          return updated
            .filter((e) => new Date(e.time_fired) >= intervalStart)
            .slice(0, 2000)
        })
        
        // If this is a state_changed event and we have pending events, trigger revalidation
        // to check if any pending service calls now have descendants
        if (event.type === "state_changed" && hasPendingEvents) {
          setTimeout(() => {
            revalidate()
          }, 100)
        }
      } catch (err) {
        console.error("Failed to parse WS message", err)
      }
    }

    // Subscribe to socket messages using the new subscriber pattern
    const unsubscribe = onSocketMessage(handleMessage)

    // Cleanup: unsubscribe from messages
    return unsubscribe
  }, [mutateEvents, selectedInterval.hours, revalidate, hasPendingEvents])

  // Summary stats (excluding time_changed events from display)
  const displayEvents = events.filter((e) => e.type !== "time_changed")
  const total = displayEvents.length
  const stateChanged = displayEvents.filter((e) => e.type === "state_changed").length
  const callService = displayEvents.filter((e) => e.type === "call_service").length
  const domainSpecific = displayEvents.filter((e) => e.type === "domain_specific").length
  const domainSpecificTypes = useMemo(() => {
    const types = new Set<string>()
    displayEvents
      .filter((e) => e.type === "domain_specific")
      .forEach((e) => {
        const data = e.data as DomainSpecificEventData
        if (data?.type) {
          types.add(data.type)
        }
      })
    return types.size
  }, [displayEvents])

  // Event type options (excluding time_changed)
  const eventTypeOptions = useMemo(() => {
    const types = new Set<EventType>(events.filter((e) => e.type !== "time_changed").map((e) => e.type))
    return [
      { id: "all", name: "All Types" },
      ...Array.from(types).map((t) => ({
        id: t,
        name: t
          .split("_")
          .map((word) => capitalize(word))
          .join(" "),
      })),
    ]
  }, [events])

  // Filtered events (by type, always exclude time_changed)
  const filteredEvents = useMemo(() => {
    let filtered = events.filter((e) => e.type !== "time_changed")
    
    // Filter by type
    if (typeFilter !== "all") {
      filtered = filtered.filter((e) => e.type === typeFilter)
    }
    
    // Filter optimistic events if enabled
    if (hideOptimistic) {
      filtered = filtered.filter((e) => e.context?.origin !== "service")
    }
    
    return filtered
  }, [events, typeFilter, hideOptimistic])

  return (
    <div className="container py-6 space-y-8">
      {/* HEADER */}
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between pb-4">
        <h1 className="text-3xl font-semibold">Event Timeline</h1>
      </div>

      {loading ? (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-1.5">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="p-2 h-12 rounded-md w-full" />
            ))}
          </div>
          <div className="flex flex-col gap-3 pt-2">
            <div className="flex flex-wrap items-center gap-2">
              <Skeleton className="h-9 w-48 rounded-md" />
            </div>
          </div>
          <Skeleton className="h-[400px] w-full rounded-lg" />
          <div className="bg-card/50 rounded-lg overflow-x-auto mt-2">
            <Skeleton className="h-96 w-full rounded-md" />
          </div>
        </>
      ) : (
        <>
          {/* SUMMARY STATS */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-1.5">
            <SummaryStatCard label="Total Events" value={total} compact />
            <SummaryStatCard label="State Changed" value={stateChanged} compact />
            <SummaryStatCard label="Service Calls" value={callService} compact />
            <SummaryStatCard label="Domain Specific" value={domainSpecific} compact />
          </div>

          {/* EVENT TYPE CHART */}
          <EventTypeChart
            events={events}
            hours={selectedInterval.hours}
            onHoursChange={(hours) => {
              const interval = TIME_INTERVALS.find((i) => i.hours === hours)
              if (interval) {
                setTimeInterval(interval.value)
              }
            }}
          />

          {/* FILTERS */}
          <div className="flex flex-col gap-3 pt-2">
            <div className="flex flex-wrap items-center gap-2 justify-between">
              <div className="flex items-center gap-3">
                <Select value={typeFilter} onValueChange={setTypeFilter}>
                  <SelectTrigger aria-label="Event type" className="w-[200px]">
                    <SelectValue placeholder="Select event type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectLabel>Event Types</SelectLabel>
                      {eventTypeOptions.map((opt) => (
                        <SelectItem key={opt.id} value={opt.id}>
                          {opt.name}
                        </SelectItem>
                      ))}
                    </SelectGroup>
                  </SelectContent>
                </Select>
                
                {/* Hide optimistic toggle */}
                <div className="flex items-center gap-2">
                  <Switch 
                    id="hide-optimistic" 
                    checked={hideOptimistic} 
                    onCheckedChange={setHideOptimistic}
                  />
                  <Label htmlFor="hide-optimistic" className="text-sm text-muted-foreground cursor-pointer">
                    Hide optimistic state updates
                  </Label>
                </div>
              </div>
              
              {/* Search input on the right */}
              <div className="relative w-full md:w-[300px]">
                <SearchIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
                <Input
                  type="text"
                  placeholder="Search events..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-9 pr-9"
                />
                {searchQuery && (
                  <button
                    type="button"
                    onClick={() => setSearchQuery("")}
                    className="absolute right-3 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                    aria-label="Clear search"
                  >
                    <XIcon className="h-4 w-4" />
                  </button>
                )}
              </div>
            </div>
          </div>

          {/* TABLE */}
          <div className="bg-card/50 rounded-lg overflow-x-auto mt-2">
            <DataTable columns={createColumns(events)} data={filteredEvents} allEvents={events} searchQuery={searchQuery} hideOptimistic={hideOptimistic} />
          </div>
        </>
      )}
    </div>
  )
}
