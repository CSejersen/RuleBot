
"use client"

import { Event } from "@/types/events"
import { Badge } from "@/components/ui/badge"
import { StateChangedDetails } from "./state-changed-details"

interface EventDetailProps {
  event: Event
}

export function EventDetails({ event }: EventDetailProps) {
  // Choose the specific detail component based on event type
  const renderTypeDetail = () => {
    switch (event.type) {
      case "state_changed":
        return <StateChangedDetails event={event} />
      default:
        return <p className="text-muted-foreground">No renderer for this event type</p>
    }
  }

  return (
    <div className="space-y-4 text-sm">
      {/* Shared header */}
      <div className="grid grid-cols-2 gap-2">
        <p><strong>Event Type:</strong> <Badge variant="outline">{event.type}</Badge></p>
        <p><strong>Time Fired:</strong> {new Date(event.time_fired).toLocaleString()}</p>
        <p><strong>ID:</strong> {event.context_id ?? "—"}</p>
      </div>

      {/* Type-specific content */}
      {renderTypeDetail()}
    </div>
  )
}
