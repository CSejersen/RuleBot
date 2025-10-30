"use client"

import { Event } from "@/types/events"
import { Badge } from "@/components/ui/badge"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { ENTITY_STATE_KEY_MAP } from "@/lib/entity-display-map"

interface StateChangedDetailsProps {
  event: Event
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

  // Prepare main state display
  let readableOld, readableNew
  if (mainChanged) {
    if (typeof newState.state === "boolean" && map) {
      readableOld = oldState?.state ? map.trueLabel : map.falseLabel
      readableNew = newState.state ? map.trueLabel : map.falseLabel
    } else {
      readableOld = oldState?.state
      readableNew = newState.state
    }
  }

  // Collect attribute diffs
  const oldAttrs = oldState?.attributes || {}
  const newAttrs = newState?.attributes || {}
  const diffs: [string, any][] = []

  for (const [key, newVal] of Object.entries(newAttrs)) {
    const oldVal = oldAttrs[key]
    if (JSON.stringify(oldVal) !== JSON.stringify(newVal)) {
      diffs.push([key, newVal])
    }
  }

  return (
    <div className="space-y-4">
      {/* Main state change only if it changed */}
      {mainChanged && (
        <div className="border-b border-muted pb-2">
          <h3 className="font-semibold">State Change</h3>
          <p>{entityId}: {readableOld ?? "?"} → {readableNew ?? "?"}</p>
        </div>
      )}

      {/* Changed attributes */}
      {diffs.length > 0 ? (
        <div className="border-b border-muted pb-2">
          <h3 className="font-semibold">Changed Attributes</h3>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Attribute</TableHead>
                <TableHead>Old Value</TableHead>
                <TableHead>New Value</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {diffs.map(([key, newVal]) => (
                <TableRow key={key}>
                  <TableCell>{key}</TableCell>
                  <TableCell>{JSON.stringify(oldAttrs[key])}</TableCell>
                  <TableCell>{JSON.stringify(newVal)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      ) : (
        !mainChanged && <p className="text-muted-foreground">No changes detected</p>
      )}
    </div>
  )
}
