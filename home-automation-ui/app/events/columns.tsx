"use client"

import { ColumnDef } from "@tanstack/react-table"
import { Event } from "@/types/events"
import { Badge } from "@/components/ui/badge"
import { ENTITY_ICON_MAP, ENTITY_STATE_KEY_MAP } from "@/lib/entity-display-map"

// helper to format timestamps
function formatTime(iso: string) {
  const date = new Date(iso)
  return date.toLocaleTimeString("en-GB", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  })
}

// helper for event summary of state_changed events
function summarizeStateChanged(event: Event) {
  const d = event.data as any
  if (!d) return "Unknown state change"

  const entityId = d.entity_id || d.new_state?.entity_id
  const oldState = d.old_state
  const newState = d.new_state
  if (!entityId || !newState) return "Incomplete state data"

  const oldMain = oldState?.state
  const newMain = newState?.state

  const entityType = entityId.split(".")[0]
  const map = ENTITY_STATE_KEY_MAP[entityType]

  // Case 1: main state changed
  if (oldMain !== newMain) {
    let readableOld = oldMain
    let readableNew = newMain

    if (typeof newMain === "boolean" && map) {
      readableOld = oldMain ? map.trueLabel : map.falseLabel
      readableNew = newMain ? map.trueLabel : map.falseLabel
    }

    return `${entityId}: ${readableOld ?? "?"} → ${readableNew ?? "?"}`
  }

  // Case 2: main state same → detect changed attribute(s)
  const oldAttrs = oldState?.attributes || {}
  const newAttrs = newState?.attributes || {}

  const diffs: string[] = []
  for (const [key, newVal] of Object.entries(newAttrs)) {
    const oldVal = oldAttrs[key]
    if (JSON.stringify(oldVal) !== JSON.stringify(newVal)) {
      diffs.push(`${key} ${oldVal ?? "?"} → ${newVal ?? "?"}`)
    }
  }

  if (diffs.length === 0) return `${entityId}: updated (no visible change)`
  if (diffs.length === 1) return `${entityId}: ${diffs[0]}`

  const preview = diffs.slice(0, 2).join(", ")
  return `${entityId}: ${preview}${diffs.length > 2 ? "…" : ""}`
}

// helper to create a human-readable summary
function renderSummary(event: Event) {
  switch (event.type) {
    case "state_changed": {
      const summary = summarizeStateChanged(event)
      return (
        <div className="flex flex-col">
          <div className="flex items-center space-x-2">
            <Badge variant="outline" className="text-xs">state_changed</Badge>
            <span className="font-medium truncate">{summary}</span>
          </div>
        </div>
      )
    }

    case "call_service": {
      const d = event.data as any
      const domain = d?.domain
      const service = d?.service
      const entity = d?.entity_id
      return (
        <div className="flex flex-col">
          <div className="flex items-center space-x-2">
            <Badge variant="outline" className="text-xs">call_service</Badge>
            <span className="font-medium truncate">{`${domain}.${service}`}</span>
          </div>
          {entity && (
            <span className="text-muted-foreground text-xs truncate">
              Target: {entity}
            </span>
          )}
        </div>
      )
    }

    case "time_changed": {
      return (
        <div className="flex items-center space-x-2">
          <Badge variant="outline" className="text-xs">time_changed</Badge>
          <span className="text-muted-foreground text-xs">Clock tick</span>
        </div>
      )
    }

    default:
      return (
        <div className="flex items-center space-x-2">
          <Badge variant="secondary" className="text-xs">{event.type}</Badge>
          <span className="text-muted-foreground text-xs">Unrecognized event</span>
        </div>
      )
  }
}

export const columns: ColumnDef<Event>[] = [
  {
    accessorKey: "time_fired",
    header: "Time",
    cell: ({ row }) => {
      const iso = row.getValue("time_fired") as string
      return (
        <div className="text-sm text-muted-foreground font-mono">
          {formatTime(iso)}
        </div>
      )
    },
  },
  {
    id: "summary",
    header: "Event",
    cell: ({ row }) => renderSummary(row.original),
  },
]
