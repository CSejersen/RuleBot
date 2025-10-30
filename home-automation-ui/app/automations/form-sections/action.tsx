"use client"

import { useEffect, useState } from "react"
import { Action, Automation } from "@/types/automation"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { FormItem, FormLabel, FormControl } from "@/components/ui/form"
import { EntitySelector } from "@/components/selectors/entity-selector"
import { ServiceSelector } from "@/components/selectors/service-selector"
import { Trash2 } from "lucide-react"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Badge } from "@/components/ui/badge"
import { useFieldArray, useFormContext } from "react-hook-form"

interface ServiceSpec {
  name: string
  required_params: Record<
    string,
    {
      DataType: string
      Description: string
    }
  >
  allowed_targets: {
    Type: string
    EntityTypes?: string[]
  }[]
}

export function ActionSection() {
  const { control, watch, setValue } = useFormContext<Automation>()
  const { fields, append, remove, update } = useFieldArray({ control, name: "actions" })
  const [openAction, setOpenAction] = useState<string | undefined>(undefined)

  const [services, setServices] = useState<ServiceSpec[]>([])

  useEffect(() => {
    fetch("/api/services")
      .then((res) => res.json())
      .then((data) => {
        const normalized = data.services.map((s: any) => ({
          ...s,
          allowed_targets: Array.isArray(s.allowed_targets) ? s.allowed_targets : [s.allowed_targets],
        }))
        setServices(normalized)
      })
  }, [])

  const addAction = () => {
    const newAction: Action = { service: "", targets: [], params: {} }
    const newIndex = fields.length
    append(newAction)
    setOpenAction(`action-${newIndex}`)
  }

  const updateAction = (index: number, updatedAction: Action) => {
    update(index, updatedAction)
  }

  const removeAction = (index: number) => {
    remove(index)
    if (openAction === `action-${index}`) {
      setOpenAction(undefined)
    } else if (openAction && parseInt(openAction.split("-")[1]) > index) {
      // Update open action index if an action before it was removed
      const currentIndex = parseInt(openAction.split("-")[1])
      setOpenAction(`action-${currentIndex - 1}`)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-base font-semibold">Actions</h3>
        <Button size="sm" onClick={addAction}>Add Action</Button>
      </div>

      {fields.length === 0 && (
        <div className="border rounded-md p-4 text-sm text-muted-foreground mt-8">
          No actions yet. Add an action to define what should happen when this automation runs.
        </div>
      )}

      <Accordion type="single" collapsible className="w-full" value={openAction} onValueChange={setOpenAction}>
        {fields.map((field, index) => {
          const action = watch(`actions.${index}`) as Action
          const selectedService = services.find((s) => s.name === action.service)
          const allowsEntities =
            selectedService?.allowed_targets.some(
              (t) => Array.isArray(t.Type) && t.Type.includes("entity")
            ) ?? false
          const summary = action.service || ""
          return (
            <div key={field.id} className="rounded-md border mb-2">
              <AccordionItem value={`action-${index}`} className="px-2">
                <AccordionTrigger className="pr-4">
                  <div className="flex items-center justify-between gap-3 w-full text-left">
                    <div className="flex items-center gap-2 min-w-0">
                      <Badge variant="secondary" className="capitalize">Call service</Badge>
                      <span className="truncate max-w-[320px]">{summary}</span>
                    </div>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className="h-8 w-8 p-0 text-muted-foreground hover:text-red-600 hover:bg-red-50 shrink-0"
                      onClick={(e) => {
                        e.stopPropagation()
                        removeAction(index)
                      }}
                      aria-label="Remove action"
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  </div>
                </AccordionTrigger>
                <AccordionContent>
                  <div className="space-y-6 p-3 pb-4">
                    {/* Service selector */}
                    <FormItem>
                      <FormLabel>Service</FormLabel>
                      <FormControl>
                        <ServiceSelector
                          value={action.service}
                          onChange={(val) => {
                            const newService = services.find((s) => s.name === val)
                            const newAllowsEntities = newService?.allowed_targets.some((t) => t.Type === "entity") ?? false
                            updateAction(index, {
                              service: val,
                              targets: newAllowsEntities ? [{ entity_id: "" }] : [],
                              params: {},
                            })
                          }}
                        />
                      </FormControl>
                    </FormItem>

                    {/* Targets section */}
                    {allowsEntities && (
                      <TargetsEditor
                        action={action}
                        index={index}
                        selectedService={selectedService}
                        updateAction={updateAction}
                      />
                    )}

                    {/* Params section */}
                    {selectedService && Object.keys(selectedService.required_params).length > 0 && (
                      <>
                        <h4 className="text-sm font-medium text-gray-700 mt-4">Parameters</h4>
                        {Object.entries(selectedService.required_params).map(([paramName, paramSpec]) => (
                          <FormItem key={paramName}>
                            <FormLabel>
                              {paramName} ({paramSpec.DataType})
                            </FormLabel>
                            <FormControl>
                              <Input
                                type="text"
                                value={action.params?.[paramName] ?? ""}
                                onChange={(e) =>
                                  setValue(
                                    `actions.${index}.params.${paramName}` as const,
                                    e.target.value,
                                  )
                                }
                              />
                            </FormControl>
                            <p className="text-sm text-muted-foreground">{paramSpec.Description}</p>
                          </FormItem>
                        ))}
                      </>
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

function EnsureFirstTarget({
  action,
  index,
  ensuresEnabled,
  updateAction,
}: {
  action: Action
  index: number
  ensuresEnabled: boolean
  updateAction: (index: number, updatedAction: Action) => void
}) {
  useEffect(() => {
    if (!ensuresEnabled) return
    if ((action.targets?.length ?? 0) === 0) {
      updateAction(index, { ...action, targets: [{ entity_id: "" }] })
    }
  }, [ensuresEnabled, action.targets?.length])
  return null
}

function TargetsEditor({
  action,
  index,
  selectedService,
  updateAction,
}: {
  action: Action
  index: number
  selectedService?: ServiceSpec
  updateAction: (index: number, updatedAction: Action) => void
}) {
  const allowedEntityTypes = selectedService?.allowed_targets
    .find((t) => Array.isArray(t.Type) && t.Type.includes("entity"))
    ?.EntityTypes

  const addOne = () => {
    updateAction(index, { ...action, targets: [...action.targets, { entity_id: "" }] })
  }

  const removeAt = (targetIndex: number) => {
    const newTargets = action.targets.filter((_, i) => i !== targetIndex)
    updateAction(index, { ...action, targets: newTargets })
  }

  return (
    <div className="space-y-2">
      <EnsureFirstTarget action={action} index={index} ensuresEnabled={true} updateAction={updateAction} />
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium text-gray-700">Targets</h4>
        <div className="flex items-center gap-2">
          <Button size="sm" variant="outline" onClick={addOne}>Add Target</Button>
        </div>
      </div>

      {action.targets.map((target, targetIndex) => (
        <div key={targetIndex} className="flex items-center space-x-2">
          <EntitySelector
            value={target.entity_id!}
            allowedEntityTypes={allowedEntityTypes}
            onChange={(val) => {
              const newTargets = [...action.targets]
              newTargets[targetIndex] = { entity_id: val }
              updateAction(index, { ...action, targets: newTargets })
            }}
            onlyEnabled={true}
          />
          {targetIndex > 0 && (
            <button
              type="button"
              className="p-1 rounded hover:bg-red-50 text-red-600 hover:text-red-700 cursor-pointer"
              onClick={() => removeAt(targetIndex)}
              aria-label="Remove target"
              title="Remove target"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          )}
        </div>
      ))}
    </div>
  )
}
