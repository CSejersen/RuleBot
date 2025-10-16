"use client"

import { ColumnDef } from "@tanstack/react-table"
import { Event } from "@/app/api/events/route"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { ChevronDown } from "lucide-react"

type MultiSelectFilterProps<T extends string | number> = {
  column: any
  table: any
  valueType?: "string" | "number" // optional, defaults to string
}

function MultiSelectFilter<T extends string | number>({
  column,
  table,
  valueType = "string",
}: MultiSelectFilterProps<T>) {
  // Extract unique values
  const uniqueValues = Array.from(
    new Set(
      table.getPreFilteredRowModel().rows.map((row: any) => {
        const value = row.getValue(column.id)

        // If valueType is number AND value is an array, use its length
        const numericValue =
          valueType === "number"
            ? Array.isArray(value)
              ? value.length
              : Number(value)
            : String(value)

        return numericValue
      })
    )
  ).filter(v => v !== "" && v != null) as T[]

  const selectedValues = (column.getFilterValue() as T[]) ?? []

  const toggleValue = (value: T, checked: boolean) => {
    const newValues = checked
      ? [...selectedValues, value]
      : selectedValues.filter(v => v !== value)

    column.setFilterValue(newValues.length ? newValues : undefined)
  }

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" className="p-1 text-xs">
          <ChevronDown className="h-4 w-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-48">
        <div className="flex flex-col space-y-1 max-h-64 overflow-auto">
          {uniqueValues.map(value => (
            <label key={value} className="flex items-center space-x-2">
              <Checkbox
                checked={selectedValues.includes(value)}
                onCheckedChange={checked => toggleValue(value, !!checked)}
              />
              <span>{value}</span>
            </label>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  )
}


export const columns: ColumnDef<Event>[] = [
  {
    accessorKey: "source",
    header: ({ column, table }) => (
      <div className="flex items-center space-x-1">
        <span className="text-sm font-medium">Source</span>
        <MultiSelectFilter column={column} table={table} />
      </div>
    ),
    filterFn: (row, columnId, filterValues: string[]) => {
      if (!filterValues || filterValues.length === 0) return true
      return filterValues.includes(row.getValue(columnId))
    },
    cell: ({ row }) => <div>{row.getValue("source")}</div>,
  },
  {
    accessorKey: "type",
    header: ({ column, table }) => (
      <div className="flex items-center space-x-1">
        <span className="text-sm font-medium">Type</span>
        <MultiSelectFilter column={column} table={table} />
      </div>
    ),
    filterFn: (row, columnId, filterValues: string[]) => {
      if (!filterValues || filterValues.length === 0) return true
      return filterValues.includes(row.getValue(columnId))
    },
    cell: ({ row }) => (
      <div className="text-sm text-muted-foreground">{row.getValue("type")}</div>
    ),
  },
  {
    accessorKey: "entity",
    header: ({ column, table }) => (
      <div className="flex items-center space-x-1">
        <span className="text-sm font-medium">Entity</span>
        <MultiSelectFilter column={column} table={table} />
      </div>
    ),
    filterFn: (row, columnId, filterValues: string[]) => {
      if (!filterValues || filterValues.length === 0) return true
      return filterValues.includes(row.getValue(columnId))
    },
    cell: ({ row }) => (
      <div className="text-sm text-muted-foreground">{row.getValue("entity")}</div>
    ),
  },
  {
    accessorKey: "stateChange",
    header: ({ column, table }) => (
      <div className="flex items-center space-x-1">
        <span className="text-sm font-medium">State Change</span>
        <MultiSelectFilter column={column} table={table} />
      </div>
    ),
    filterFn: (row, columnId, filterValues: string[]) => {
      if (!filterValues || filterValues.length === 0) return true
      return filterValues.includes(row.getValue(columnId))
    },
    cell: ({ row }) => (
      <div className="text-sm text-muted-foreground">{row.getValue("stateChange")}</div>
    ),
  },
  {
    accessorKey: "timestamp",
    header: "Time",
    cell: ({ row }) => {
      const isoTime = row.getValue("timestamp") as string
      const date = new Date(isoTime)
      const formatted = date.toLocaleString("en-GB", {
        year: "numeric",
        month: "short",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        hour12: false,
      })
      return <div className="text-sm text-muted-foreground">{formatted}</div>
    },
  },
  {
    accessorKey: "triggeredRules",
    header: ({ column, table }) => (
      <div className="flex items-center space-x-1">
        <span className="text-sm font-medium">Triggered Rules</span>
        <MultiSelectFilter<number> column={column} table={table} valueType="number" />
      </div>
    ),
    filterFn: (row, columnId, filterValues: number[]) => {
      if (!filterValues || filterValues.length === 0) return true
      const length = (row.getValue(columnId) as string[]).length
      return filterValues.includes(length)
    },
    cell: ({ row }) => {
      const rules = row.getValue("triggeredRules") as string[]
      return (
        <div className="text-left text-sm text-muted-foreground">
          {rules.length}
        </div>
      )
    },
  },
]
