import { Event, DomainSpecificEventData } from "@/types/events"
import { ENTITY_STATE_KEY_MAP } from "@/lib/entity-display-map"

type DescendantWithDepth = {
  event: Event
  depth: number
}

/**
 * Resolves the ancestor chain (events that caused this event)
 * by recursively following parent_id relationships
 */
export function resolveAncestorChain(event: Event, allEvents: Event[]): Event[] {
  const chain: Event[] = []
  const visited = new Set<string>() // Prevent cycles
  
  let currentEvent: Event | undefined = event
  
  while (currentEvent?.context?.parent_id) {
    const parentId: string = currentEvent.context.parent_id
    
    // Prevent infinite loops
    if (visited.has(parentId)) {
      break
    }
    visited.add(parentId)
    
    // Find the parent event (where context.id matches parent_id)
    const parent: Event | undefined = allEvents.find(
      (e) => e.context?.id === parentId
    )
    
    if (parent) {
      chain.unshift(parent) // Add to beginning to maintain chronological order
      currentEvent = parent
    } else {
      // Parent not found in loaded events
      break
    }
  }
  
  return chain
}

/**
 * Resolves the descendant chain (events that were caused by this event)
 * by finding all events where their context.parent_id matches this event's context.id
 */
export function resolveDescendantChain(event: Event, allEvents: Event[]): Event[] {
  if (!event.context?.id) {
    return []
  }
  
  const descendants: DescendantWithDepth[] = []
  const visited = new Set<string>() // Prevent cycles
  
  function collectDescendants(parentContextId: string, depth: number) {
    if (visited.has(parentContextId)) {
      return
    }
    visited.add(parentContextId)
    
    // Find all events where their parent_id matches the parent context id
    const children = allEvents.filter(
      (e) => e.context?.parent_id === parentContextId
    )
    
    for (const child of children) {
      descendants.push({ event: child, depth })
      
      // Recursively collect descendants of this child
      if (child.context?.id) {
        collectDescendants(child.context.id, depth + 1)
      }
    }
  }
  
  collectDescendants(event.context.id, 1)
  
  // Sort by depth first (parents before children), then timestamp, then event type
  const sorted = descendants.sort((a, b) => {
    // First: sort by depth
    if (a.depth !== b.depth) {
      return a.depth - b.depth
    }
    
    // Second: sort by timestamp
    const timeDiff = new Date(a.event.time_fired).getTime() - new Date(b.event.time_fired).getTime()
    if (timeDiff !== 0) {
      return timeDiff
    }
    
    // Third: sort by event type (call_service before state_changed)
    const typePriority = (type: string): number => {
      switch (type) {
        case "call_service":
          return 0
        case "state_changed":
          return 1
        default:
          return 2
      }
    }
    
    return typePriority(a.event.type) - typePriority(b.event.type)
  })
  
  // Return just the events (strip depth metadata)
  return sorted.map(item => item.event)
}

/**
 * Groups events by their parent context ID
 * Returns an array of groups, where each group contains events with the same parent
 * Groups are returned in the order they were first encountered in the input array
 */
export function groupEventsByParent(events: Event[]): Event[][] {
  const groups: Map<string, Event[]> = new Map()
  const groupOrder: string[] = []
  
  for (const event of events) {
    const parentId = event.context?.parent_id || ""
    if (!groups.has(parentId)) {
      groups.set(parentId, [])
      groupOrder.push(parentId)
    }
    groups.get(parentId)!.push(event)
  }
  
  // Return groups in the order they were first encountered
  return groupOrder.map(parentId => groups.get(parentId)!)
}

/**
 * Gets a brief summary string for an event (for display in the chain)
 */
export function getEventSummary(event: Event): string {
  switch (event.type) {
    case "state_changed":
      const data = event.data as { entity_id?: string }
      return data?.entity_id || "State Changed"
    case "call_service":
      const serviceData = event.data as { domain?: string; service?: string; entity_id?: string }
      if (serviceData?.domain && serviceData?.service) {
        return `${serviceData.domain}.${serviceData.service}`
      }
      return "Service Call"
    case "time_changed":
      return "Time Changed"
    case "domain_specific":
      const domainData = event.data as DomainSpecificEventData
      return domainData?.type || "Domain Specific"
    default:
      return event.type
  }
}

/**
 * Gets detailed information about what changed in an event
 */
export function getEventDetails(event: Event): string {
  if (event.type === "call_service") {
    const d = event.data as { entity_id?: string }
    return d?.entity_id || event.context?.id || event.id
  }
  
  if (event.type === "domain_specific") {
    const d = event.data as DomainSpecificEventData
    if (!d?.type) {
      return event.context?.id || event.id
    }
    
    const eventData = d.data || {}
    
    // Build description with all simple fields
    const fieldParts: string[] = []
    
    for (const [key, value] of Object.entries(eventData)) {
      // Skip null/undefined values
      if (value === null || value === undefined) {
        continue
      }
      
      // Check if value is a simple type
      const valueType = typeof value
      if (valueType === "string" || valueType === "number" || valueType === "boolean") {
        // Format the value
        let formattedValue: string
        if (valueType === "boolean") {
          formattedValue = value ? "true" : "false"
        } else {
          formattedValue = String(value)
        }
        fieldParts.push(`${key}: ${formattedValue}`)
      } else if (valueType === "object") {
        // For objects/arrays, just show "object"
        fieldParts.push(`${key}: object`)
      }
    }
    
    if (fieldParts.length === 0) {
      return "no data"
    }
    
    // Format: "type: field1: value1, field2: value2, ..."
    return `${fieldParts.join(", ")}`
  }
  
  if (event.type !== "state_changed") {
    // For other non-state-changed events, return the ID
    return event.context?.id || event.id
  }

  const d = event.data as any
  if (!d) return event.context?.id || event.id

  const entityId = d.entity_id || d.new_state?.entity_id
  const oldState = d.old_state
  const newState = d.new_state
  if (!newState) return event.context?.id || event.id

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

    return `${readableOld ?? "?"} → ${readableNew ?? "?"}`
  }

  // Case 2: main state same → detect changed attribute(s)
  const oldAttrs = oldState?.attributes || {}
  const newAttrs = newState?.attributes || {}

  const diffs: string[] = []
  for (const [key, newVal] of Object.entries(newAttrs)) {
    const oldVal = oldAttrs[key]
    if (JSON.stringify(oldVal) !== JSON.stringify(newVal)) {
      diffs.push(`${key}: ${oldVal ?? "?"} → ${newVal ?? "?"}`)
    }
  }

  if (diffs.length === 0) return "(no visible change)"
  if (diffs.length === 1) return diffs[0]

  // Multiple attributes changed - show first one + count
  const remainingCount = diffs.length - 1
  return `${diffs[0]} (+${remainingCount} more)`
}

/**
 * Checks if an event has any descendants (events caused by this event)
 */
export function hasDescendants(event: Event, allEvents: Event[]): boolean {
  if (!event.context?.id) {
    return false
  }
  const descendants = resolveDescendantChain(event, allEvents)
  return descendants.length > 0
}

/**
 * Determines if a service call is pending (waiting for state update)
 * Returns true if:
 * - Event is a call_service type
 * - Event has no descendants
 * - Event is < 5 seconds old
 */
export function isPending(event: Event, allEvents: Event[]): boolean {
  // Only check call_service events
  if (event.type !== "call_service") {
    return false
  }
  
  // Must have no descendants
  if (hasDescendants(event, allEvents)) {
    return false
  }
  
  // Must be recent (< 5 seconds) to be considered pending
  const eventTime = new Date(event.time_fired).getTime()
  const now = Date.now()
  const ageSeconds = (now - eventTime) / 1000
  
  return ageSeconds >= 0 && ageSeconds < 5
}

/**
 * Determines if a service call is likely aggregated (combined with other calls)
 * Returns true if:
 * - Event is a call_service type
 * - Event has no descendants
 * - Event is old enough (>= 5 seconds) to rule out "pending" state
 */
export function isLikelyAggregated(event: Event, allEvents: Event[]): boolean {
  // Only check call_service events
  if (event.type !== "call_service") {
    return false
  }
  
  // Must have no descendants
  if (hasDescendants(event, allEvents)) {
    return false
  }
  
  // Must be old enough (>= 5 seconds) to rule out pending state
  const eventTime = new Date(event.time_fired).getTime()
  const now = Date.now()
  const ageSeconds = (now - eventTime) / 1000
  
  return ageSeconds >= 5
}

/**
 * Gets the status of a service call event
 * Returns status based on event type, descendants, and age
 */
export function getServiceCallStatus(
  event: Event, 
  allEvents: Event[]
): "pending" | "aggregated" | "has_descendants" | "not_service_call" {
  // Only check call_service events
  if (event.type !== "call_service") {
    return "not_service_call"
  }
  
  // Check if event has descendants
  if (hasDescendants(event, allEvents)) {
    return "has_descendants"
  }
  
  // Calculate age
  const eventTime = new Date(event.time_fired).getTime()
  const now = Date.now()
  const ageSeconds = (now - eventTime) / 1000
  
  // Return status based on age
  if (ageSeconds >= 0 && ageSeconds < 5) {
    return "pending"
  } else if (ageSeconds >= 5) {
    return "aggregated"
  }
  
  // Edge case: event is in the future (shouldn't happen, but handle gracefully)
  return "aggregated"
}

