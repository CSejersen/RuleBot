"use client"

import { Badge } from "@/components/ui/badge"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"

export function AggregatedBadge() {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Badge 
          variant="outline" 
          className="text-xs bg-muted/50 dark:bg-muted/50 border-muted-foreground/40 dark:border-muted-foreground/35 text-muted-foreground/90 dark:text-muted-foreground/85 opacity-90"
        >
          Aggregated
        </Badge>
      </TooltipTrigger>
      <TooltipContent>
        <p className="max-w-xs">
          This service call was likely aggregated with other calls. 
        </p>
      </TooltipContent>
    </Tooltip>
  )
}

