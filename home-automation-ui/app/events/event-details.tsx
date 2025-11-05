
"use client"

import { Event } from "@/types/events"
import { Badge } from "@/components/ui/badge"
import { StateChangedDetails } from "./state-changed-details"
import { ServiceCallDetails } from "./service-call-details"
import { DomainSpecificEventDetails } from "./domain-specific-event-details"
import { CausalityChain } from "./causality-chain"
import { getTypeBadgeVariant } from "./columns"

interface EventDetailProps {
  event: Event
  allEvents: Event[]
  onEventChange?: (event: Event) => void
  hideOptimistic?: boolean
}

export function EventDetails({ event, allEvents, onEventChange, hideOptimistic = false }: EventDetailProps) {
  // Choose the specific detail component based on event type
  const renderTypeDetail = () => {
    switch (event.type) {
      case "state_changed":
        return <StateChangedDetails event={event} />
      case "call_service":
        return <ServiceCallDetails event={event} />
      case "domain_specific":
        return <DomainSpecificEventDetails event={event} />
      default:
        return <p className="text-muted-foreground">No renderer for this event type</p>
    }
  }

  return (
    <div className="space-y-4 text-sm">
      {/* Shared header */}
      <div className="space-y-2 pb-3 border-b border-muted">
        <div className="flex items-start gap-3">
          <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide w-20 flex-shrink-0">
            Type
          </div>
          <div className="flex-1 min-w-0">
            <Badge variant={getTypeBadgeVariant(event.type)}>{event.type}</Badge>
          </div>
        </div>
        <div className="flex items-start gap-3">
          <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide w-20 flex-shrink-0">
            Time
          </div>
          <div className="flex-1 min-w-0">
            {new Date(event.time_fired).toLocaleString()}
          </div>
        </div>
        {event.context?.origin && (
          <div className="flex items-start gap-3">
            <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide w-20 flex-shrink-0">
              Origin
            </div>
            <div className="flex-1 min-w-0">
              {event.context.origin}
            </div>
          </div>
        )}
      </div>

      {/* Type-specific content */}
      {renderTypeDetail()}

      {/* Causality chain */}
      <CausalityChain event={event} allEvents={allEvents} onEventClick={onEventChange} hideOptimistic={hideOptimistic} />
    </div>
  )
}
