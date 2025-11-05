"use client"

import { ColumnDef } from "@tanstack/react-table"
import { Automation } from "@/types/automation/automation"
import { Switch } from "@/components/ui/switch"

export function createColumns(onEnableChange: (automation: Automation, enabled: boolean) => void): ColumnDef<Automation>[] {
  return [
    {
      accessorKey: "alias",
      header: "Alias",
    },
    {
      accessorKey: "last_triggered",
      header: "Last Triggered",
      cell: ({ row }) => {
        const isoTime = row.getValue("last_triggered") as string | null
        if (!isoTime) return <div className="text-sm text-muted-foreground">N/A</div>

        // Treat MySQL DATETIME (no timezone) as UTC
        const normalized = isoTime.endsWith("Z") ? isoTime : isoTime + "Z"
        const date = new Date(normalized)

        if (isNaN(date.getTime())) {
          return <div className="text-sm text-muted-foreground">Invalid date</div>
        }

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
      accessorKey: "enabled",
      header: "Enabled",
      cell: ({ row }) => {
        const automation = row.original
        const value = row.getValue("enabled")
        const isEnabled = value === 1 || value === true || value === "1"

        return (
          <Switch
            checked={isEnabled}
            onCheckedChange={(checked) => onEnableChange(automation, checked)}
            onClick={(e) => e.stopPropagation()}
            aria-label={`Toggle ${automation.alias}`}
          />
        )
      },
    },
  ]
}
