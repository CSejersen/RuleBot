"use client"

import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"

import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  useReactTable,
} from "@tanstack/react-table"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog"
import { useState } from "react"

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[]
  data: TData[]
}

export function DataTable<TData extends Record<string, any>, TValue>({
  columns,
  data,
}: DataTableProps<TData, TValue>) {
  const [columnFilters, setColumnFilters] = useState<any[]>([])
  const [selectedEvent, setSelectedEvent] = useState<TData | null>(null)
  const [isOpen, setIsOpen] = useState(false)

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    state: { columnFilters },
    onColumnFiltersChange: setColumnFilters,
    initialState: { pagination: { pageSize: 17 } },
  })

  return (
    <div>
      <div className="overflow-hidden rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && "selected"}
                  className="cursor-pointer hover:bg-muted/50 transition-colors"
                  onClick={() => {
                    setSelectedEvent(row.original)
                    setIsOpen(true)
                  }}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  No results.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <div className="flex items-center justify-between">
        <Button
          variant="outline"
          size="sm"
          onClick={() => setColumnFilters([])}
          disabled={columnFilters.length === 0} // optional: gray out if no filters
        >
          Reset Filters
        </Button>
        <div className="flex items-center justify-end space-x-2 py-4">
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            Previous
          </Button>
          <span className="text-sm">
            Page {table.getState().pagination.pageIndex + 1} of {table.getPageCount()}
          </span>
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            Next
          </Button>
        </div>
      </div>


      {/* Event Detail Dialog */}
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Event Details</DialogTitle>
            <DialogDescription>
              {selectedEvent
                ? `Information for event with ID: ${selectedEvent.id}`
                : "Information about the selected event."}
            </DialogDescription>
          </DialogHeader>

          {selectedEvent ? (
            <div className="space-y-4 text-sm">
              <div className="grid grid-cols-2 gap-2">
                <p><strong>Source:</strong> {selectedEvent.source}</p>
                <p><strong>Type:</strong> {selectedEvent.type}</p>
                <p><strong>Entity:</strong> {selectedEvent.entity}</p>
                <p><strong>State Change:</strong> {selectedEvent.stateChange}</p>
                <p><strong>Timestamp:</strong> {selectedEvent.timestamp ?? "—"}</p>
              </div>

              {selectedEvent.triggeredRules.length > 0 && (
                <div className="space-y-1">
                  <p className="font-semibold mb-1">Rules Triggered:</p>
                  {selectedEvent.triggeredRules.map((rule: string, index: number) => (
                    <Badge key={index} variant="secondary" className="text-xs">
                      {rule}
                    </Badge>
                  ))}
                </div>
              )}

              <div>
                <p className="font-semibold mb-1">Payload:</p>
                <pre className="bg-muted p-3 rounded-md overflow-auto text-xs max-h-60">
                  {JSON.stringify(selectedEvent.payload, null, 2)}
                </pre>
              </div>
            </div>
          ) : (
            <p>No event selected.</p>
          )}
        </DialogContent>
      </Dialog>
    </div>
  )
}
