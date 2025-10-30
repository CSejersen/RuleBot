"use client"

import { useState, useEffect, useMemo } from "react"
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
import { Check, ChevronsUpDown } from "lucide-react"
import { cn } from "@/lib/utils"
import { State } from "@/types/state"

interface AttributeSelectorProps {
  entityId: string
  value?: string
  onChange: (attribute: string) => void
}

export function AttributeSelector({ entityId, value, onChange }: AttributeSelectorProps) {
  const [attributes, setAttributes] = useState<string[]>([])
  const [open, setOpen] = useState(false)

  // Fetch attributes when entityId changes
  useEffect(() => {
    if (!entityId) {
      setAttributes([])
      return
    }

    const url = `/api/states?entity_id=${encodeURIComponent(entityId)}`

    fetch(url)
      .then((res) => res.json())
      .then((state: State) => {
        const attrs = Object.keys(state.attributes || {})
        setAttributes(attrs)
      })
      .catch(console.error)
  }, [entityId])

  // Compute label for trigger
  const triggerLabel = useMemo(() => (value ? value : "Select Attribute"), [value])

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="secondary"
          size="sm"
          aria-expanded={open}
          className="inline-flex items-center gap-1"
          disabled={!entityId}
        >
          {triggerLabel}
          <ChevronsUpDown className="h-4 w-4 opacity-60" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[--radix-popover-trigger-width] p-0">
        <Command>
          <CommandInput placeholder="Search attribute..." />
          <CommandList
            className="max-h-60 overflow-y-auto overscroll-contain"
            style={{ WebkitOverflowScrolling: "touch" }}
            onWheelCapture={(e) => e.stopPropagation()}
          >
            <CommandEmpty>No attributes found.</CommandEmpty>
            <CommandGroup>
              {attributes.map((attr) => (
                <CommandItem
                  key={attr}
                  value={attr}
                  onSelect={(currentValue) => {
                    onChange(currentValue)
                    setOpen(false)
                  }}
                >
                  <Check
                    className={cn(
                      "mr-2 h-4 w-4",
                      value === attr ? "opacity-100" : "opacity-0"
                    )}
                  />
                  {attr}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
