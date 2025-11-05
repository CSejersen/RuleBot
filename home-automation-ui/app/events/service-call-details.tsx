"use client"

import { Event } from "@/types/events"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"

interface ServiceCallDetailsProps {
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

export function ServiceCallDetails({ event }: ServiceCallDetailsProps) {
  const d = event.data as any
  if (!d?.domain || !d?.service) {
    return <p className="text-muted-foreground">Incomplete service call data</p>
  }

  const domain = d.domain
  const service = d.service
  const entityId = d.entity_id
  const serviceData = d.service_data || {}

  const serviceDataEntries = Object.entries(serviceData)

  return (
    <div className="space-y-4">
      {/* Service Info */}
      <div className="space-y-2 pb-3 border-b border-muted">
        <div className="flex items-start gap-3">
          <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide w-20 flex-shrink-0">
            Domain
          </div>
          <div className="flex-1 min-w-0">
            <code className="text-sm font-mono">{domain}</code>
          </div>
        </div>
        <div className="flex items-start gap-3">
          <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide w-20 flex-shrink-0">
            Service
          </div>
          <div className="flex-1 min-w-0">
            <code className="text-sm font-mono">{service}</code>
          </div>
        </div>
        {entityId && (
          <div className="flex items-start gap-3">
            <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide w-20 flex-shrink-0">
              Target
            </div>
            <div className="flex-1 min-w-0">
              <code className="text-sm font-mono">{entityId}</code>
            </div>
          </div>
        )}
      </div>

      {/* Service Data */}
      {serviceDataEntries.length > 0 ? (
        <div className="space-y-2">
          <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wide">
            Parameters ({serviceDataEntries.length})
          </h4>
          <div className="rounded-md border border-muted overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow className="h-8">
                  <TableHead className="h-8 py-1 text-xs font-medium w-[120px]">Parameter</TableHead>
                  <TableHead className="h-8 py-1 text-xs font-medium">Value</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {serviceDataEntries.map(([key, value]) => {
                  const complex = isComplexValue(value)
                  
                  return (
                    <TableRow key={key} className="h-auto">
                      <TableCell className="py-1.5 text-xs font-medium">{key}</TableCell>
                      <TableCell className="py-1.5 text-xs">
                        {complex ? (
                          <pre className="text-xs bg-muted p-1.5 rounded max-w-xs overflow-auto max-h-32">
                            {formatValue(value)}
                          </pre>
                        ) : (
                          <code className="text-xs bg-muted px-1 py-0.5 rounded">
                            {formatValue(value)}
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
      ) : (
        <div className="py-3 text-center text-sm text-muted-foreground border-t border-muted">
          No service parameters provided
        </div>
      )}
    </div>
  )
}
