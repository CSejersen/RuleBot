"use client"

import { Badge } from "@/components/ui/badge"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"

export function PendingBadge() {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Badge
          variant="outline"
          className="text-xs bg-orange-50 dark:bg-orange-900/20 border-orange-200 dark:border-orange-800 text-orange-700 dark:text-orange-300"
        >
          Pending
        </Badge>
      </TooltipTrigger>
      <TooltipContent>
        <p className="max-w-xs">
          This service call is still waiting for a state update.
        </p>
      </TooltipContent>
    </Tooltip>
  )
}

