"use client"

import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
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
import type { ServiceSpec } from "@/types/service"

interface ServiceSelectorProps {
  value: string
  onChange: (serviceName: string) => void
}

export function ServiceSelector({ value, onChange }: ServiceSelectorProps) {
  const [services, setServices] = useState<ServiceSpec[]>([])
  const [open, setOpen] = useState(false)

  useEffect(() => {
    fetch("/api/services")
      .then((res) => res.json())
      .then((data: { services: ServiceSpec[] }) => setServices(data.services || []))
      .catch(console.error)
  }, [])

  

  return (
    <div className="space-y-2">
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            type="button"
            variant="secondary"
            size="sm"
            aria-expanded={open}
            className="inline-flex items-center gap-1"
          >
            {value ? value : "Select Service"}
            <ChevronsUpDown className="h-4 w-4 opacity-60" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[--radix-popover-trigger-width] p-0">
          <Command>
            <CommandInput placeholder="Search service..." />
            <CommandList
              className="max-h-60 overflow-y-auto overscroll-contain"
              style={{ WebkitOverflowScrolling: "touch" }}
              onWheelCapture={(e) => e.stopPropagation()}
            >
              <CommandEmpty>No services found.</CommandEmpty>
              <CommandGroup>
                {services.map((s) => (
                  <CommandItem
                    key={s.name}
                    value={s.name}
                    onSelect={(currentValue) => {
                    onChange(currentValue)
                      setOpen(false)
                    }}
                  >
                    <Check
                      className={cn(
                        "mr-2 h-4 w-4",
                        value === s.name ? "opacity-100" : "opacity-0"
                      )}
                    />
                    {s.name}
                  </CommandItem>
                ))}
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
    </div>
  )
}
