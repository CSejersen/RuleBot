"use client"

import { Automation, BaseTrigger, StateTrigger, EventTrigger } from "@/types/automation"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { FormItem, FormLabel, FormControl, FormDescription } from "@/components/ui/form"
import { EntitySelector } from "@/components/selectors/entity-selector"
import { AttributeSelector } from "@/components/selectors/attribute-selector"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Badge } from "@/components/ui/badge"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { useFieldArray, useFormContext } from "react-hook-form"
import { useEffect, useState } from "react"
import { Trash2, ChevronDown, HelpCircle } from "lucide-react"
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip"

export function TriggerSection() {
  const { control, setValue, watch } = useFormContext<Automation>()
  const { fields, append, remove, update } = useFieldArray({ control, name: "triggers" })
  const [openTrigger, setOpenTrigger] = useState<string | undefined>(undefined)

  const addTrigger = (type: BaseTrigger["type"]) => {
    const newTrigger: BaseTrigger =
      type === "state"
        ? { type: "state", data: { entity_id: "", attribute: undefined, from: undefined, to: undefined } }
        : { type: "event", data: { event_type: "" } }
    const newIndex = fields.length
    append(newTrigger)
    setOpenTrigger(`trigger-${newIndex}`)
  }

  const updateTrigger = (index: number, updatedTrigger: BaseTrigger) => {
    update(index, updatedTrigger)
  }

  const removeTrigger = (index: number) => {
    remove(index)
    if (openTrigger === `trigger-${index}`) {
      setOpenTrigger(undefined)
    } else if (openTrigger && parseInt(openTrigger.split("-")[1]) > index) {
      // Update open trigger index if a trigger before it was removed
      const currentIndex = parseInt(openTrigger.split("-")[1])
      setOpenTrigger(`trigger-${currentIndex - 1}`)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-base font-semibold">Triggers</h3>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button size="sm" variant="default" className="whitespace-nowrap inline-flex items-center gap-1">
              <span>Add trigger</span>
              <ChevronDown className="h-3 w-3 opacity-70" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {TRIGGER_TYPE_OPTIONS.map((opt) => (
              <DropdownMenuItem key={opt.value} onClick={() => addTrigger(opt.value)}>
                <div className="flex flex-col">
                  <span className="font-medium">{opt.label}</span>
                  <span className="text-xs text-muted-foreground">{opt.description}</span>
                </div>
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {fields.length === 0 && (
        <div className="border rounded-md p-4 text-sm text-muted-foreground mt-8">
            No triggers yet. Add a trigger to decide when this automation should run.
        </div>
      )}

      <Accordion type="single" collapsible className="w-full" value={openTrigger} onValueChange={setOpenTrigger}>
        {fields.map((field, index) => {
          const trigger = watch(`triggers.${index}`) as BaseTrigger
          const summary = !trigger.type
            ? "Select trigger type"
            : trigger.type === "state"
              ? (trigger.data as StateTrigger).entity_id || ""
              : (trigger.data as EventTrigger).event_type || ""
          return (
            <div key={field.id} className="rounded-md border mb-2">
              <AccordionItem value={`trigger-${index}`} className="px-2">
                <AccordionTrigger className="pr-4">
                  <div className="flex items-center justify-between gap-3 w-full text-left">
                    <div className="flex items-center gap-2 min-w-0">
                      <Badge variant="secondary" className="capitalize">{trigger.type}</Badge>
                      <span className="truncate max-w-[320px]">
                        {summary}
                      </span>
                    </div>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className="h-8 w-8 p-0 text-muted-foreground hover:text-red-600 hover:bg-red-50 shrink-0"
                      onClick={(e) => {
                        e.stopPropagation()
                        removeTrigger(index)
                      }}
                      aria-label="Remove trigger"
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  </div>
                </AccordionTrigger>
                <AccordionContent>
                  <div className="space-y-4 p-3 pb-4">
                    {trigger.type === "state" && (
                      <StateTriggerForm
                        trigger={trigger.data as StateTrigger}
                        onChange={(data) => updateTrigger(index, { ...trigger, data })}
                      />
                    )}

                    {trigger.type === "event" && (
                      <EventTriggerForm
                        trigger={trigger.data as EventTrigger}
                        onChange={(data) => updateTrigger(index, { ...trigger, data })}
                      />
                    )}
                  </div>
                </AccordionContent>
              </AccordionItem>
            </div>
          )
        })}
      </Accordion>
    </div>
  )
}

// Trigger type dropdown
// (Trigger type control was moved to the Accordion header)
const TRIGGER_TYPE_OPTIONS: Array<{ value: BaseTrigger["type"], label: string, description: string }> = [
  { value: "state", label: "State Change", description: "React to entity state or attribute changes" },
  { value: "event", label: "Event Type", description: "React to a specific event type" },
]


interface StateTriggerFormProps {
  trigger: StateTrigger
  onChange: (data: StateTrigger) => void
}

export function StateTriggerForm({ trigger, onChange }: StateTriggerFormProps) {
  const [hasAttributes, setHasAttributes] = useState(false)

  useEffect(() => {
    let cancelled = false
    async function load() {
      if (!trigger.entity_id) {
        if (!cancelled) setHasAttributes(false)
        return
      }
      try {
        const res = await fetch(`/api/states?entity_id=${encodeURIComponent(trigger.entity_id)}`)
        const state = await res.json()
        const attrs = Object.keys(state?.attributes || {})
        if (!cancelled) setHasAttributes(attrs.length > 0)
      } catch {
        if (!cancelled) setHasAttributes(false)
      }
    }
    load()
    return () => {
      cancelled = true
    }
  }, [trigger.entity_id])
  return (
    <div className="space-y-4">
      {/* Entity ID */}
      <FormItem>
        <FormLabel>Entity ID</FormLabel>
        <FormControl>
          <div className="inline-flex">
            <EntitySelector
              value={trigger.entity_id}
              onChange={(val) => onChange({ ...trigger, entity_id: val })}
              onlyEnabled={true}
            />
          </div>
        </FormControl>
        <FormDescription>
          Select the entity whose state changes should trigger this automation.
        </FormDescription>
      </FormItem>

      {/* Attribute */}
      {hasAttributes && (
        <FormItem>
          <FormLabel className="flex items-center gap-1">
            <span>Attribute <span className="text-xs text-muted-foreground">(optional)</span></span>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <HelpCircle className="h-3.5 w-3.5 text-muted-foreground" />
                </TooltipTrigger>
                <TooltipContent side="top" align="center" className="max-w-xs">
                  By default, this trigger watches the entity’s main state. Select an attribute to trigger on that instead.
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </FormLabel>
          <FormControl>
            <div className="inline-flex">
              <AttributeSelector
                entityId={trigger.entity_id}
                value={trigger.attribute}
                onChange={(attr) => onChange({ ...trigger, attribute: attr })}
              />
            </div>
          </FormControl>
        </FormItem>
      )}

      {/* From */}
      <FormItem>
        <FormLabel className="flex items-center gap-1">
          <span>From <span className="text-xs text-muted-foreground">(optional)</span></span>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <HelpCircle className="h-3.5 w-3.5 text-muted-foreground" />
              </TooltipTrigger>
              <TooltipContent side="top" align="center" className="max-w-xs">
                Only trigger when the value changes from this state.
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </FormLabel>
        <FormControl>
          <Input
            value={trigger.from || ""}
            onChange={(e) => onChange({ ...trigger, from: e.target.value || undefined })}
          />
        </FormControl>
      </FormItem>

      {/* To */}
      <FormItem>
        <FormLabel className="flex items-center gap-1">
          <span>To <span className="text-xs text-muted-foreground">(optional)</span></span>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <HelpCircle className="h-3.5 w-3.5 text-muted-foreground" />
              </TooltipTrigger>
              <TooltipContent side="top" align="center" className="max-w-xs">
                Only trigger when the value changes to this state.
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </FormLabel>
        <FormControl>
          <Input
            value={trigger.to || ""}
            onChange={(e) => onChange({ ...trigger, to: e.target.value || undefined })}
          />
        </FormControl>
      </FormItem>
    </div>
  )
}

// EventTrigger form
interface EventTriggerFormProps {
  trigger: EventTrigger
  onChange: (data: EventTrigger) => void
}
function EventTriggerForm({ trigger, onChange }: EventTriggerFormProps) {
  return (
    <FormItem>
      <FormLabel>Event Type</FormLabel>
      <FormControl>
        <Input
          type="text"
          placeholder="e.g. state_changed, button_pressed"
          value={trigger.event_type}
          onChange={(e) => onChange({ ...trigger, event_type: e.target.value })}
        />
      </FormControl>
    </FormItem>
  )
}
