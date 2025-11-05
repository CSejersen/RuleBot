"use client"

import { Event } from "@/types/events"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { ENTITY_STATE_KEY_MAP } from "@/lib/entity-display-map"
import { ArrowRight } from "lucide-react"

interface StateChangedDetailsProps {
  event: Event
}

// Helper to format attribute values nicely
function formatValue(value: any): string {
  if (value === null || value === undefined) {
    return "—"
  }
  if (typeof value === "boolean") {
    return value ? "true" : "false"
  }
  if (typeof value === "number") {
    return String(value)
  }
  if (typeof value === "string") {
    return value
  }
  if (typeof value === "object") {
    // Check if it's a date string or object
    if (value instanceof Date || (typeof value === "string" && !isNaN(Date.parse(value)))) {
      try {
        return new Date(value).toLocaleString()
      } catch {
        return String(value)
      }
    }
    // Format objects/arrays nicely
    try {
      return JSON.stringify(value, null, 2)
    } catch {
      return String(value)
    }
  }
  return String(value)
}

// Helper to check if a value is complex (object/array)
function isComplexValue(value: any): boolean {
  return typeof value === "object" && value !== null && !(value instanceof Date)
}

export function StateChangedDetails({ event }: StateChangedDetailsProps) {
  const d = event.data as any
  if (!d?.new_state) return <p className="text-muted-foreground">Incomplete state data</p>

  const entityId = d.entity_id || d.new_state?.entity_id
  const oldState = d.old_state
  const newState = d.new_state

  const entityType = entityId?.split(".")[0]
  const map = ENTITY_STATE_KEY_MAP[entityType]

  const mainChanged = oldState?.state !== newState?.state
  const hasOldState = oldState !== null && oldState !== undefined

  // Prepare main state display
  let readableOld, readableNew
  if (mainChanged || hasOldState) {
    if (typeof newState.state === "boolean" && map) {
      readableOld = oldState?.state !== undefined 
        ? (oldState.state ? map.trueLabel : map.falseLabel)
        : "?"
      readableNew = newState.state ? map.trueLabel : map.falseLabel
    } else {
      readableOld = oldState?.state !== undefined ? formatValue(oldState.state) : "?"
      readableNew = formatValue(newState.state)
    }
  }

  // Collect attribute diffs
  const oldAttrs = oldState?.attributes || {}
  const newAttrs = newState?.attributes || {}
  const diffs: [string, any, any][] = []

  // Check all attributes in new state
  for (const [key, newVal] of Object.entries(newAttrs)) {
    const oldVal = oldAttrs[key]
    if (JSON.stringify(oldVal) !== JSON.stringify(newVal)) {
      diffs.push([key, oldVal, newVal])
    }
  }

  // Also check for attributes that were removed
  for (const [key, oldVal] of Object.entries(oldAttrs)) {
    if (!(key in newAttrs)) {
      diffs.push([key, oldVal, undefined])
    }
  }

  return (
    <div className="space-y-4">
      {/* Entity Info */}
      <div className="space-y-2 pb-3 border-b border-muted">
        <div className="flex items-start gap-3">
          <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide w-20 flex-shrink-0">
            Entity
          </div>
          <div className="flex-1 min-w-0">
            <code className="text-sm font-mono truncate">{entityId}</code>
          </div>
        </div>
        {mainChanged && (
          <div className="flex items-start gap-3">
            <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide w-20 flex-shrink-0">
              State
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 flex-wrap">
                <code className="text-sm font-mono text-muted-foreground">{readableOld}</code>
                <ArrowRight className="h-3 w-3 text-muted-foreground flex-shrink-0" />
                <code className="text-sm font-mono">{readableNew}</code>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Changed Attributes */}
      {diffs.length > 0 && (
        <div className="space-y-2">
          <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wide">
            Changed Attributes ({diffs.length})
          </h4>
          <div className="rounded-md border border-muted overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow className="h-8">
                  <TableHead className="h-8 py-1 text-xs font-medium w-[120px]">Attribute</TableHead>
                  <TableHead className="h-8 py-1 text-xs font-medium">Old Value</TableHead>
                  <TableHead className="h-8 py-1 text-xs font-medium">New Value</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {diffs.map(([key, oldVal, newVal]) => {
                  const oldComplex = isComplexValue(oldVal)
                  const newComplex = isComplexValue(newVal)
                  
                  return (
                    <TableRow key={key} className="h-auto">
                      <TableCell className="py-1.5 text-xs font-medium">{key}</TableCell>
                      <TableCell className="py-1.5 text-xs">
                        {oldVal === undefined ? (
                          <span className="text-muted-foreground italic">—</span>
                        ) : oldComplex ? (
                          <pre className="text-xs bg-muted p-1.5 rounded max-w-xs overflow-auto max-h-32">
                            {formatValue(oldVal)}
                          </pre>
                        ) : (
                          <code className="text-xs bg-muted px-1 py-0.5 rounded">
                            {formatValue(oldVal)}
                          </code>
                        )}
                      </TableCell>
                      <TableCell className="py-1.5 text-xs">
                        {newVal === undefined ? (
                          <span className="text-muted-foreground italic">removed</span>
                        ) : newComplex ? (
                          <pre className="text-xs bg-primary/10 p-1.5 rounded max-w-xs overflow-auto max-h-32">
                            {formatValue(newVal)}
                          </pre>
                        ) : (
                          <code className="text-xs bg-primary/10 px-1 py-0.5 rounded font-medium">
                            {formatValue(newVal)}
                          </code>
                        )}
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </div>
        </div>
      )}

      {/* No Changes Message */}
      {!mainChanged && diffs.length === 0 && (
        <div className="py-3 text-center text-sm text-muted-foreground border-t border-muted">
          No changes detected in this event
        </div>
      )}
    </div>
  )
}
