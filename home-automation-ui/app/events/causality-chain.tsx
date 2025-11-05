"use client"

import { Event } from "@/types/events"
import { Badge } from "@/components/ui/badge"
import { resolveAncestorChain, resolveDescendantChain, getEventSummary, groupEventsByParent, getEventDetails, getServiceCallStatus } from "./causality-chain-utils"
import { getTypeBadgeVariant } from "./columns"
import { ArrowDown, ArrowUp } from "lucide-react"
import { OptimisticBadge } from "@/components/common/optimistic-badge"
import { AggregatedBadge } from "@/components/common/aggregated-badge"
import { PendingBadge } from "@/components/common/pending-badge"
import { useServiceCallStatus } from "./use-service-call-status"

interface CausalityChainProps {
  event: Event
  allEvents: Event[]
  onEventClick?: (event: Event) => void
  hideOptimistic?: boolean
}

export function CausalityChain({ event, allEvents, onEventClick, hideOptimistic = false }: CausalityChainProps) {
  const ancestors = resolveAncestorChain(event, allEvents)
  const descendants = resolveDescendantChain(event, allEvents)
  
  // Filter optimistic events if enabled
  const filteredAncestors = hideOptimistic ? ancestors.filter(e => e.context?.origin !== "service") : ancestors
  const filteredDescendants = hideOptimistic ? descendants.filter(e => e.context?.origin !== "service") : descendants

  // Use hook for dynamic status updates (only for service calls)
  const { status: serviceCallStatus } = useServiceCallStatus(event, allEvents)
  const isPending = serviceCallStatus === "pending"
  const isAggregated = serviceCallStatus === "aggregated"

  // Don't render if no chain exists and event is not pending/aggregated
  if (filteredAncestors.length === 0 && filteredDescendants.length === 0 && !isPending && !isAggregated) {
    return null
  }

  const getBorderColorClass = (parentId: string): string => {
    const colors = [
      'border-l-blue-500',
      'border-l-green-500', 
      'border-l-purple-500',
      'border-l-orange-500'
    ]
    // Simple hash function to consistently map parent ID to color
    let hash = 0
    for (let i = 0; i < parentId.length; i++) {
      const char = parentId.charCodeAt(i)
      hash = ((hash << 5) - hash) + char
      hash = hash & hash // Convert to 32bit integer
    }
    return colors[Math.abs(hash) % colors.length]
  }

  const renderEventItem = (chainEvent: Event, isCurrent: boolean = false) => {
    const summary = getEventSummary(chainEvent)
    const isClickable = !isCurrent && onEventClick !== undefined
    const isOptimistic = chainEvent.context?.origin === "service"
    
    return (
      <div
        key={chainEvent.id}
        className={`flex items-start gap-3 p-2 rounded-lg border ${
          isCurrent
            ? "bg-primary/10 border-primary"
            : isOptimistic
            ? "bg-yellow-50 dark:bg-yellow-900/10 border-yellow-200 dark:border-yellow-800"
            : "bg-card/50 border-border"
        } ${
          isClickable ? "cursor-pointer hover:bg-muted/50 transition-colors" : ""
        }`}
        onClick={isClickable ? () => onEventClick?.(chainEvent) : undefined}
      >
        <div className="flex flex-col items-center gap-1 min-w-[60px]">
          <Badge variant={getTypeBadgeVariant(chainEvent.type)} className="text-xs">
            {chainEvent.type}
          </Badge>
          <span className="text-xs text-muted-foreground">
            {new Date(chainEvent.time_fired).toLocaleTimeString()}
          </span>
        </div>
        <div className="flex-1">
          <div className="flex items-center gap-2">
            <p className="text-sm font-medium">{summary}</p>
            {isOptimistic && <OptimisticBadge />}
          </div>
          <p className="text-xs text-muted-foreground mt-1 truncate">
            {getEventDetails(chainEvent)}
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4 mt-6 border-t pt-4">
      <h3 className="text-sm font-semibold text-foreground">Causality Chain</h3>

      {/* Ancestors (causes) */}
      {filteredAncestors.length > 0 && (
        <div className="space-y-2">
          <div className="space-y-2 ml-4">
            {filteredAncestors.map((ancestor) => (
              <div key={ancestor.id}>
                {renderEventItem(ancestor)}
              </div>
            ))}
          </div>
          <div className="flex items-center gap-2 text-xs font-medium text-muted-foreground uppercase tracking-wide">
            <ArrowUp className="h-3 w-3" />
            <span>Caused By ({filteredAncestors.length} ancestor{filteredAncestors.length !== 1 ? 's' : ''})</span>
          </div>
        </div>
      )}

      {/* Current Event */}
      <div className="space-y-2">
        {renderEventItem(event, true)}
      </div>

      {/* Descendants (effects) */}
      {filteredDescendants.length > 0 && (
        <>
          <div className="flex items-center gap-2 text-xs font-medium text-muted-foreground uppercase tracking-wide">
            <ArrowDown className="h-3 w-3" />
            <span>Caused ({filteredDescendants.length} descendant{filteredDescendants.length !== 1 ? 's' : ''})</span>
          </div>
          {(() => {
            // Group descendants by parent to visually group siblings
            const groupedDescendants = groupEventsByParent(filteredDescendants)
            
            return (
              <div className="space-y-2">
                <div className="space-y-2 ml-4">
                  {groupedDescendants.map((group, groupIndex) => {
                    const parentId = group[0]?.context?.parent_id || ""
                    return group.length === 1 ? (
                      <div key={groupIndex}>
                        {renderEventItem(group[0])}
                      </div>
                    ) : (
                      <div key={groupIndex} className={`border-l-4 ${getBorderColorClass(parentId)} pl-3 space-y-2`}>
                        {group.map((descendant) => (
                          <div key={descendant.id}>
                            {renderEventItem(descendant)}
                          </div>
                        ))}
                      </div>
                    )
                  })}
                </div>
              </div>
            )
          })()}
        </>
      )}

      {/* Show pending/aggregated message when service call has no descendants */}
      {(isPending || isAggregated) && filteredDescendants.length === 0 && (
        <div className="space-y-2">
          <div className="flex items-center gap-2 text-xs font-medium text-muted-foreground uppercase tracking-wide">
            <ArrowDown className="h-3 w-3" />
            <span>Caused</span>
          </div>
          {isPending ? (
            <div className="ml-4 p-3 rounded-lg border border-orange-200 dark:border-orange-800 bg-orange-50/50 dark:bg-orange-900/10">
              <div className="flex items-start gap-2">
                <PendingBadge />
                <div className="flex-1">
                  <p className="text-sm text-muted-foreground">
                    Waiting for state update... This service call was sent recently and may still be processed by the device.
                  </p>
                </div>
              </div>
            </div>
          ) : (
            <div className="ml-4 p-3 rounded-lg border border-muted-foreground/30 dark:border-muted-foreground/25 bg-muted/50 dark:bg-muted/40 opacity-85">
              <div className="flex items-start gap-2">
                <AggregatedBadge />
                <div className="flex-1">
                  <p className="text-sm text-muted-foreground/85">
                    No state changes detected. This service call was likely aggregated with other rapid calls.
                  </p>
                </div>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

