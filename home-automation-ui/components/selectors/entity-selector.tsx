"use client"

import { useEffect, useMemo, useState } from "react"
import { Button } from "@/components/ui/button"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import { ChevronsUpDown, Check } from "lucide-react"
import { cn } from "@/lib/utils"
import { Entity } from "@/types/entity"

interface EntitySelectorProps {
  value: string
  onChange: (entityId: string) => void
  allowedEntityTypes?: string[]
  onlyEnabled?: boolean
}

export function EntitySelector({ value, onChange, allowedEntityTypes, onlyEnabled }: EntitySelectorProps) {
  const [entities, setEntities] = useState<Entity[]>([])
  const [open, setOpen] = useState(false)

  // Fetch entities and apply static filters from props
  useEffect(() => {
    fetch("/api/entities")
      .then((res) => res.json())
      .then((data: { entities: Entity[] }) => {
        let all = data.entities || []
        if (allowedEntityTypes && allowedEntityTypes.length > 0) {
          all = all.filter((e) => allowedEntityTypes.includes(e.type))
        }
        if (onlyEnabled) {
          all = all.filter((e) => e.enabled)
        }
        setEntities(all)
      })
      .catch(console.error)
  }, [allowedEntityTypes, onlyEnabled])

  // Compute label for trigger
  const triggerLabel = useMemo(() => (value ? value : "Select Entity"), [value])

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="secondary"
          size="sm"
          aria-expanded={open}
          className="inline-flex items-center gap-1"
        >
          {triggerLabel}
          <ChevronsUpDown className="h-4 w-4 opacity-60" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[--radix-popover-trigger-width] p-0">
        <Command>
          <CommandInput placeholder="Search entity..." />
            <CommandList
              className="max-h-60 overflow-y-auto overscroll-contain"
              style={{ WebkitOverflowScrolling: "touch" }}
              onWheelCapture={(e) => e.stopPropagation()}
            >
              <CommandEmpty>No entities found.</CommandEmpty>
              <CommandGroup>
                {entities.map((e) => (
                  <CommandItem
                    key={e.external_id || e.entity_id}
                    value={e.entity_id}
                    onSelect={(currentValue) => {
                      onChange(currentValue)
                      setOpen(false)
                    }}
                  >
                    <Check
                      className={cn(
                        "mr-2 h-4 w-4",
                        value === e.entity_id ? "opacity-100" : "opacity-0"
                      )}
                    />
                    {e.entity_id}
                  </CommandItem>
                ))}
              </CommandGroup>
            </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
