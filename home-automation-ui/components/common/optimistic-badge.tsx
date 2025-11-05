"use client"

import { Badge } from "@/components/ui/badge"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"

export function OptimisticBadge() {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Badge 
          variant="outline" 
          className="text-xs bg-yellow-100 dark:bg-yellow-900/20 border-yellow-300 dark:border-yellow-800"
        >
          Optimistic
        </Badge>
      </TooltipTrigger>
      <TooltipContent>
        <p className="max-w-xs">
          This event was published by a service without receiving a state-update for the entity. Usually emitted by services that target entities which do not report their own state back after updates.
        </p>
      </TooltipContent>
    </Tooltip>
  )
}

