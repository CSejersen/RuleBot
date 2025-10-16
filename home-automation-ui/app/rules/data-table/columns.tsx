"use client"

import { ColumnDef } from "@tanstack/react-table"
import { Rule } from "@/app/api/rules/types/rule"

export const columns: ColumnDef<Rule>[] = [
  {
    accessorKey: "alias",
    header: "Alias",
  },
  {
    accessorKey: "lastTriggered",
    header: "Last Triggered",
    cell: ({ row }) => {
      // Make sure we use the right key
      const isoTime = row.getValue("lastTriggered") as string;
      if (!isoTime) return <div className="text-sm text-muted-foreground">N/A</div>;

      const date = new Date(isoTime);
      if (isNaN(date.getTime())) {
        return <div className="text-sm text-muted-foreground">Invalid date</div>;
      }

      const formatted = date.toLocaleString("en-GB", {
        year: "numeric",
        month: "short",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        hour12: false,
      });

      return <div className="text-sm text-muted-foreground">{formatted}</div>;
    },
  },
  {
    accessorKey: "active",
    header: "Active",
  }
]
