"use client"

import { Button } from "@/components/ui/button"
import { ColumnDef, flexRender, getCoreRowModel, getFilteredRowModel, getPaginationRowModel, useReactTable } from "@tanstack/react-table"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useState, useMemo } from "react"
import { Event } from "@/types/events"
import { EventDetails } from "./event-details"
import Fuse from "fuse.js"
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react"
import { getServiceCallStatus } from "./causality-chain-utils"

interface DataTableProps<TData extends Event, TValue> {
  columns: ColumnDef<TData, TValue>[]
  data: TData[]
  allEvents: TData[]
  searchQuery?: string
  hideOptimistic?: boolean
}

// Helper to get searchable text from an event
function getSearchableText(event: Event): string {
  const parts: string[] = []
  
  // Add event type
  parts.push(event.type)
  
  // Add event ID
  parts.push(String(event.id))
  
  // Add context ID if available
  if (event.context_id) {
    parts.push(event.context_id)
  }
  
  // Extract relevant data based on event type
  const data = event.data as any
  if (data) {
    if (event.type === "state_changed") {
      const entityId = data.entity_id || data.new_state?.entity_id || data.old_state?.entity_id
      if (entityId) parts.push(entityId)
      
      const oldState = data.old_state?.state
      const newState = data.new_state?.state
      if (oldState) parts.push(String(oldState))
      if (newState) parts.push(String(newState))
      
      // Add attribute keys and values
      const attrs = data.new_state?.attributes || {}
      Object.entries(attrs).forEach(([key, value]) => {
        parts.push(key)
        if (value !== null && value !== undefined) {
          parts.push(String(value))
        }
      })
      
      // Also add old state attributes for comparison
      const oldAttrs = data.old_state?.attributes || {}
      Object.entries(oldAttrs).forEach(([key, value]) => {
        parts.push(key)
        if (value !== null && value !== undefined) {
          parts.push(String(value))
        }
      })
    } else if (event.type === "call_service") {
      if (data.domain) parts.push(data.domain)
      if (data.service) parts.push(data.service)
      if (data.entity_id) {
        const entityIds = Array.isArray(data.entity_id) ? data.entity_id : [data.entity_id]
        entityIds.forEach((id: any) => parts.push(String(id)))
      }
      
      // Add service data keys and values
      if (data.service_data) {
        Object.entries(data.service_data).forEach(([key, value]) => {
          parts.push(key)
          if (value !== null && value !== undefined) {
            if (typeof value === "object" && !Array.isArray(value)) {
              Object.entries(value as any).forEach(([subKey, subValue]) => {
                parts.push(subKey)
                if (subValue !== null && subValue !== undefined) {
                  parts.push(String(subValue))
                }
              })
            } else {
              parts.push(String(value))
            }
          }
        })
      }
    }
    
    // Add any other data as stringified values (flattened)
    const stringifyData = (obj: any, prefix = ""): void => {
      if (obj && typeof obj === "object" && !Array.isArray(obj)) {
        Object.entries(obj).forEach(([key, value]) => {
          const fullKey = prefix ? `${prefix}.${key}` : key
          if (value !== null && value !== undefined) {
            if (typeof value === "object" && !Array.isArray(value)) {
              stringifyData(value, fullKey)
            } else {
              parts.push(fullKey)
              parts.push(String(value))
            }
          }
        })
      } else if (Array.isArray(obj)) {
        obj.forEach((item) => {
          if (typeof item === "string" || typeof item === "number") {
            parts.push(String(item))
          }
        })
      }
    }
    
    // Only stringify if we haven't already extracted main fields
    if (event.type !== "state_changed" && event.type !== "call_service") {
      stringifyData(data)
    }
  }
  
  return parts.join(" ").toLowerCase()
}

// Helper function to generate page numbers with ellipsis
function getPageNumbers(currentPage: number, totalPages: number): (number | "ellipsis")[] {
  const pages: (number | "ellipsis")[] = []
  
  if (totalPages <= 7) {
    // If 7 or fewer pages, show all
    for (let i = 1; i <= totalPages; i++) {
      pages.push(i)
    }
  } else {
    // Always show first page
    pages.push(1)
    
    if (currentPage <= 4) {
      // Near the beginning: 1 2 3 4 5 ... last
      for (let i = 2; i <= 5; i++) {
        pages.push(i)
      }
      pages.push("ellipsis")
      pages.push(totalPages)
    } else if (currentPage >= totalPages - 3) {
      // Near the end: 1 ... (last-4) (last-3) (last-2) (last-1) last
      pages.push("ellipsis")
      for (let i = totalPages - 4; i <= totalPages; i++) {
        pages.push(i)
      }
    } else {
      // In the middle: 1 ... (current-1) current (current+1) ... last
      pages.push("ellipsis")
      pages.push(currentPage - 1)
      pages.push(currentPage)
      pages.push(currentPage + 1)
      pages.push("ellipsis")
      pages.push(totalPages)
    }
  }
  
  return pages
}

export function DataTable<TData extends Event, TValue>({ columns, data, allEvents, searchQuery = "", hideOptimistic = false }: DataTableProps<TData, TValue>) {
  const [columnFilters, setColumnFilters] = useState<any[]>([])
  const [selectedEvent, setSelectedEvent] = useState<TData | null>(null)
  const [isOpen, setIsOpen] = useState(false)

  // Fuzzy search implementation
  const searchableData = useMemo(() => {
    return data.map((event) => ({
      event,
      searchableText: getSearchableText(event)
    }))
  }, [data])

  const fuse = useMemo(() => {
    return new Fuse(searchableData, {
      keys: ["searchableText"],
      threshold: 0.4, // 0 = perfect match, 1 = match anything. Lower = more strict
      ignoreLocation: true,
      includeScore: true,
      minMatchCharLength: 1,
    })
  }, [searchableData])

  const filteredData = useMemo(() => {
    if (!searchQuery.trim()) {
      return data
    }
    
    const results = fuse.search(searchQuery)
    const events = results.map(result => result.item.event)
    
    // Sort by time_fired (most recent first) to maintain chronological order
    return events.sort((a, b) => {
      const timeA = new Date(a.time_fired).getTime()
      const timeB = new Date(b.time_fired).getTime()
      return timeB - timeA // Descending order (newest first)
    })
  }, [data, searchQuery, fuse])

  const table = useReactTable({
    data: filteredData,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    state: { columnFilters },
    onColumnFiltersChange: setColumnFilters,
    initialState: { pagination: { pageSize: 20 } },
  })

  const currentPage = table.getState().pagination.pageIndex + 1
  const totalPages = table.getPageCount()
  const pageNumbers = useMemo(() => getPageNumbers(currentPage, totalPages), [currentPage, totalPages])

  return (
    <div>
      {/* Table */}
      <div className="overflow-hidden">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map(headerGroup => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map(header => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map(row => {
                const isOptimistic = row.original.context?.origin === "service"
                const serviceCallStatus = getServiceCallStatus(row.original, allEvents)
                const isAggregated = serviceCallStatus === "aggregated"
                return (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && "selected"}
                  className={`cursor-pointer transition-colors ${
                    isOptimistic
                      ? "bg-yellow-50/50 dark:bg-yellow-900/5 hover:bg-yellow-50 dark:hover:bg-yellow-900/10"
                      : isAggregated
                      ? "bg-muted/40 dark:bg-muted/40 opacity-80 hover:bg-muted/70 dark:hover:bg-muted/60"
                      : "hover:bg-muted/50"
                  }`}
                  onClick={() => {
                    setSelectedEvent(row.original)
                    setIsOpen(true)
                  }}
                >
                  {row.getVisibleCells().map(cell => (
                    <TableCell key={cell.id}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
                )
              })
            ) : (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  <span className="text-muted-foreground">No events match your filter.</span>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      <div className="flex flex-col gap-4 mt-4">
        {/* Pagination Controls */}
        <div className="flex items-center justify-center gap-2">
          {/* First Page */}
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.setPageIndex(0)}
            disabled={!table.getCanPreviousPage()}
            className="h-8 w-8 p-0"
            aria-label="First page"
          >
            <ChevronsLeft className="h-4 w-4" />
          </Button>

          {/* Previous Page */}
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
            className="h-8 w-8 p-0"
            aria-label="Previous page"
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>

          {/* Page Numbers */}
          <div className="flex items-center gap-1">
            {pageNumbers.map((page, index) => {
              if (page === "ellipsis") {
                return (
                  <span key={`ellipsis-${index}`} className="px-2 text-sm text-muted-foreground">
                    ...
                  </span>
                )
              }
              return (
                <Button
                  key={page}
                  variant={page === currentPage ? "default" : "outline"}
                  size="sm"
                  onClick={() => table.setPageIndex(page - 1)}
                  className="h-8 w-8 p-0"
                  aria-label={`Page ${page}`}
                  aria-current={page === currentPage ? "page" : undefined}
                >
                  {page}
                </Button>
              )
            })}
          </div>

          {/* Next Page */}
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
            className="h-8 w-8 p-0"
            aria-label="Next page"
          >
            <ChevronRight className="h-4 w-4" />
          </Button>

          {/* Last Page */}
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.setPageIndex(table.getPageCount() - 1)}
            disabled={!table.getCanNextPage()}
            className="h-8 w-8 p-0"
            aria-label="Last page"
          >
            <ChevronsRight className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Event Detail Dialog */}
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent className="max-w-2xl max-h-[85vh] p-0 overflow-hidden">
          <DialogHeader className="flex-shrink-0 px-6 pt-6 pb-4">
            <DialogTitle>Event Details</DialogTitle>
            <DialogDescription>
              {selectedEvent ? `Details for event ID: ${selectedEvent.id}` : "Select an event to view details"}
            </DialogDescription>
          </DialogHeader>

          <ScrollArea className="max-h-[calc(85vh-120px)] px-6 pb-6">
            {selectedEvent ? (
              <EventDetails 
                event={selectedEvent} 
                allEvents={allEvents}
                onEventChange={(event) => {
                  setSelectedEvent(event as TData)
                }}
                hideOptimistic={hideOptimistic}
              />
            ) : (
              <p>No event selected.</p>
            )}
          </ScrollArea>
        </DialogContent>
      </Dialog>
    </div>
  )
}
