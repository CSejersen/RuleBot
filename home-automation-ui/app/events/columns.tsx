"use client"

import { ColumnDef } from "@tanstack/react-table"
import { Event } from "@/types/events"
import { Badge } from "@/components/ui/badge"
import { ENTITY_ICON_MAP, ENTITY_STATE_KEY_MAP } from "@/lib/entity-display-map"
import { OptimisticBadge } from "@/components/common/optimistic-badge"
import { AggregatedBadge } from "@/components/common/aggregated-badge"
import { PendingBadge } from "@/components/common/pending-badge"
import { getServiceCallStatus } from "./causality-chain-utils"

// helper to format timestamps
function formatTime(iso: string) {
  const date = new Date(iso)
  const now = new Date()
  const isToday = date.toDateString() === now.toDateString()
  const isYesterday = date.toDateString() === new Date(now.getTime() - 24 * 60 * 60 * 1000).toDateString()
  
  const timeStr = date.toLocaleTimeString("en-GB", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  })
  
  if (isToday) {
    return timeStr
  } else if (isYesterday) {
    return `Yesterday ${timeStr}`
  } else {
    const dateStr = date.toLocaleDateString("en-GB", {
      day: "2-digit",
      month: "short",
    })
    return `${dateStr} ${timeStr}`
  }
}

// helper to determine if an event has no visible changes
export function hasNoVisibleChange(event: Event): boolean {
  if (event.type !== "state_changed") {
    return false
  }

  const d = event.data as any
  if (!d) return false

  const oldState = d.old_state
  const newState = d.new_state
  if (!newState) return false

  const oldMain = oldState?.state
  const newMain = newState?.state

  // If main state changed, there's a visible change
  if (oldMain !== newMain) {
    return false
  }

  // Check for attribute changes
  const oldAttrs = oldState?.attributes || {}
  const newAttrs = newState?.attributes || {}

  for (const [key, newVal] of Object.entries(newAttrs)) {
    const oldVal = oldAttrs[key]
    if (JSON.stringify(oldVal) !== JSON.stringify(newVal)) {
      return false
    }
  }

  // No main state change and no attribute changes = no visible change
  return true
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

  if (diffs.length === 0) return `${entityId}: (no visible change)`
  if (diffs.length === 1) return `${entityId}: ${diffs[0]}`

  const preview = diffs.slice(0, 2).join(", ")
  return `${entityId}: ${preview}${diffs.length > 2 ? "…" : ""}`
}

// helper to create a human-readable description
function getDescription(event: Event): string {
  switch (event.type) {
    case "state_changed": {
      return summarizeStateChanged(event)
    }

    case "call_service": {
      const d = event.data as any
      const domain = d?.domain
      const service = d?.service
      const entity = d?.entity_id
      if (entity) {
        return `${domain}.${service} → ${entity}`
      }
      return `${domain}.${service}`
    }

    case "time_changed": {
      return "Clock tick"
    }

    case "domain_specific": {
      const d = event.data as any
      const eventType = d?.type || "domain_specific"
      const entityId = d?.entity_id
      if (entityId) {
        return `${eventType} → ${entityId}`
      }
      return eventType
    }

    default:
      return "Unrecognized event"
  }
}

// helper to get badge variant for event type
export function getTypeBadgeVariant(type: string) {
  switch (type) {
    case "state_changed":
      return "default"
    case "call_service":
      return "secondary"
    case "time_changed":
      return "outline"
    case "domain_specific":
      return "outline"
    default:
      return "outline"
  }
}

export function createColumns(allEvents: Event[]): ColumnDef<Event>[] {
  return [
    {
      accessorKey: "time_fired",
      header: "Time",
      cell: ({ row }) => {
        const iso = row.getValue("time_fired") as string
        const serviceCallStatus = getServiceCallStatus(row.original, allEvents)
        const isAggregated = serviceCallStatus === "aggregated"
        return (
          <div className={`text-sm font-mono ${isAggregated ? "text-muted-foreground/70 font-light" : "text-muted-foreground"}`}>
            {formatTime(iso)}
          </div>
        )
      },
    },
    {
      accessorKey: "type",
      header: "Event Type",
      cell: ({ row }) => {
        const event = row.original
        const type = row.getValue("type") as string
        const serviceCallStatus = getServiceCallStatus(event, allEvents)
        const isAggregated = serviceCallStatus === "aggregated"
        
        return (
          <Badge variant={getTypeBadgeVariant(type)} className={`text-xs ${isAggregated ? "opacity-60" : ""}`}>
            {type}
          </Badge>
        )
      },
    },
    {
      id: "description",
      header: "Description",
      cell: ({ row }) => {
        const description = getDescription(row.original)
        const isOptimistic = row.original.context?.origin === "service"
        const serviceCallStatus = getServiceCallStatus(row.original, allEvents)
        const isPending = serviceCallStatus === "pending"
        const isAggregated = serviceCallStatus === "aggregated"
        return (
          <div className={`text-sm max-w-[500px] ${isAggregated ? "text-muted-foreground/70 font-light" : "text-muted-foreground"}`}>
            <div className="flex items-center gap-2">
              <span className="truncate block">{description}</span>
              {isOptimistic && <OptimisticBadge />}
              {isPending && <PendingBadge />}
              {isAggregated && <AggregatedBadge />}
            </div>
          </div>
        )
      },
    },
  ]
}

// Keep the old export for backward compatibility, but it won't show aggregated badges
export const columns: ColumnDef<Event>[] = createColumns([])

