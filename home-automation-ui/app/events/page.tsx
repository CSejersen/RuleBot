"use client"

import { useEffect, useState } from "react"
import { columns } from "./columns"
import { DataTable } from "./data-table"
import { getEngineSocket } from "@/lib/engine-socket"
import { Event } from "@/types/events"

async function getData(): Promise<Event[]> {
  try {
    const res = await fetch("/api/events", { cache: "no-store" })
    if (!res.ok) {
      console.error("Failed to fetch events:", res.statusText)
      return []
    }
    const data = (await res.json()) as Event[]
    return data
  } catch (err) {
    console.error("Error fetching events:", err)
    return []
  }
}

function parseEventMessage(msg: any): Event {
  const raw = typeof msg === "string" ? JSON.parse(msg) : msg

  return {
    id: raw.id ?? raw.ID ?? 0,
    type: raw.type ?? raw.Type ?? "unknown",
    data: raw.data ?? raw.Data ?? {},
    context_id: raw.context_id ?? raw.ContextID ?? "",
    time_fired: raw.time_fired ?? raw.TimeFired ?? new Date().toISOString(),
    context: raw.context ?? raw.Context ?? null,
  }
}

export default function Dashboard() {
  const [events, setEvents] = useState<Event[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    getData()
      .then((data) => setEvents(data))
      .catch((err) => {
        console.error(err)
        setError("Failed to load events")
      })
      .finally(() => setLoading(false))

    const ws = getEngineSocket()
    ws.onmessage = (message) => {
      try {
        const event = parseEventMessage(message.data)
        setEvents((prev) => [event, ...prev])
      } catch (err) {
        console.error("Failed to parse WS message", err)
      }
    }

    return () => ws.close()
  }, [])

  if (loading) return <div>Loading events...</div>
  if (error) return <div>Error: {error}</div>

  return (
    <div className="container mx-auto py-10">
      <h2 className="text-2xl font-semibold tracking-tight mb-4">
        Event Timeline
      </h2>
      <DataTable columns={columns} data={events} />
    </div>
  )
}
