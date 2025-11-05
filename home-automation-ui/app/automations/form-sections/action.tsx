"use client"

import { useEffect, useState, useRef } from "react"
import { Action, Automation } from "@/types/automation"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { FormItem, FormLabel, FormControl, FormMessage } from "@/components/ui/form"
import { EntitySelector } from "@/components/selectors/entity-selector"
import { ServiceSelector } from "@/components/selectors/service-selector"
import { Trash2 } from "lucide-react"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Badge } from "@/components/ui/badge"
import { useFieldArray, useFormContext } from "react-hook-form"
import React from "react"
import { SaveAttemptContext } from "../create-automation-sheet"

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
  const { control, watch, setValue, formState: { errors } } = useFormContext<Automation>()
  const hasAttemptedSave = React.useContext(SaveAttemptContext)
  const { fields, append, remove, update } = useFieldArray({ control, name: "actions" })
  const [openAction, setOpenAction] = useState<string | undefined>(undefined)
  const prevFieldsLength = useRef(0)

  const [services, setServices] = useState<ServiceSpec[]>([])

  // When editing, expand the first action if it exists
  useEffect(() => {
    if (fields.length === 0) {
      // Reset when fields are cleared
      setOpenAction(undefined)
      prevFieldsLength.current = 0
    } else if (prevFieldsLength.current === 0 && fields.length > 0 && openAction === undefined) {
      // Expand first action when fields transition from empty to populated (editing mode)
      setOpenAction("action-0")
    }
    prevFieldsLength.current = fields.length
  }, [fields.length, openAction])

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
    // Don't trigger validation on update - validation will happen on save attempt
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
        <div className="space-y-2">
          <div className="border rounded-md p-4 text-sm text-muted-foreground mt-8">
            No actions yet. Add an action to define what should happen when this automation runs.
          </div>
          {hasAttemptedSave && errors.actions && (
            <p className="text-sm text-destructive mt-2">{errors.actions.message || "At least one action is required"}</p>
          )}
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
                    <div
                      className="inline-flex items-center justify-center h-8 w-8 p-0 rounded-md text-sm font-medium transition-colors text-muted-foreground hover:text-red-600 hover:bg-red-50 shrink-0 cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                      onClick={(e) => {
                        e.stopPropagation()
                        removeAction(index)
                      }}
                      onKeyDown={(e) => {
                        if (e.key === "Enter" || e.key === " ") {
                          e.preventDefault()
                          e.stopPropagation()
                          removeAction(index)
                        }
                      }}
                      role="button"
                      tabIndex={0}
                      aria-label="Remove action"
                    >
                      <Trash2 className="w-4 h-4" />
                    </div>
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
                      {hasAttemptedSave && errors.actions?.[index]?.service && (
                        <FormMessage>{(errors.actions[index] as any)?.service?.message}</FormMessage>
                      )}
                    </FormItem>

                    {/* Targets section */}
                    {allowsEntities && (
                      <TargetsEditor
                        action={action}
                        actionIndex={index}
                        selectedService={selectedService}
                        updateAction={updateAction}
                      />
                    )}

                    {/* Params section */}
                    {selectedService && Object.keys(selectedService.required_params).length > 0 && (
                      <>
                        <h4 className="text-sm font-medium text-gray-700 mt-4">Parameters</h4>
                        {Object.entries(selectedService.required_params).map(([paramName, paramSpec]) => (
                          <ParamField
                            key={paramName}
                            actionIndex={index}
                            action={action}
                            paramName={paramName}
                            paramSpec={paramSpec}
                            service={action.service}
                            setValue={setValue}
                            errors={errors}
                          />
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

interface ParamFieldProps {
  actionIndex: number
  action: Action
  paramName: string
  paramSpec: { DataType: string; Description: string }
  service: string
  setValue: (name: any, value: any) => void
  errors: any
}

function ParamField({
  actionIndex,
  action,
  paramName,
  paramSpec,
  service,
  setValue,
  errors,
}: ParamFieldProps) {
  const { setError, clearErrors } = useFormContext<Automation>()
  const hasAttemptedSave = React.useContext(SaveAttemptContext)
  const [isValidating, setIsValidating] = useState(false)
  const [validationError, setValidationError] = useState<string | null>(null)

  const validateParam = async (value: string) => {
    if (!value || !service) {
      setValidationError(null)
      clearErrors(`actions.${actionIndex}.params.${paramName}` as any)
      return
    }

    setIsValidating(true)
    setValidationError(null)

    try {
      const res = await fetch("/api/services/validate-param", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          service,
          param_name: paramName,
          value,
        }),
      })

      const data = await res.json()

      if (!res.ok) {
        const errorMsg = data.error || "Validation failed"
        setValidationError(errorMsg)
        setError(`actions.${actionIndex}.params.${paramName}` as any, {
          type: "server",
          message: errorMsg,
        })
      } else {
        setValidationError(null)
        clearErrors(`actions.${actionIndex}.params.${paramName}` as any)
      }
    } catch (error: any) {
      const errorMsg = error.message || "Failed to validate parameter"
      setValidationError(errorMsg)
      setError(`actions.${actionIndex}.params.${paramName}` as any, {
        type: "server",
        message: errorMsg,
      })
    } finally {
      setIsValidating(false)
    }
  }

  const paramValue = action.params?.[paramName] ?? ""
  const fieldError = errors.actions?.[actionIndex]?.params?.[paramName]

  return (
    <FormItem>
      <FormLabel>
        {paramName} ({paramSpec.DataType})
      </FormLabel>
      <FormControl>
        <Input
          type="text"
          value={paramValue}
          onChange={(e) => {
            const value = e.target.value
            setValue(`actions.${actionIndex}.params.${paramName}` as const, value)
            // Clear validation error on change
            if (validationError) {
              setValidationError(null)
              clearErrors(`actions.${actionIndex}.params.${paramName}` as any)
            }
          }}
          onBlur={(e) => {
            validateParam(e.target.value)
          }}
        />
      </FormControl>
      <p className="text-sm text-muted-foreground">{paramSpec.Description}</p>
      {hasAttemptedSave && (fieldError || validationError) && (
        <p className="text-sm text-destructive mt-1">
          {fieldError?.message || validationError}
        </p>
      )}
    </FormItem>
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
  }, [ensuresEnabled, action.targets?.length, action, index, updateAction])
  return null
}

function TargetsEditor({
  action,
  actionIndex,
  selectedService,
  updateAction,
}: {
  action: Action
  actionIndex: number
  selectedService?: ServiceSpec
  updateAction: (index: number, updatedAction: Action) => void
}) {
  const { formState: { errors } } = useFormContext<Automation>()
  const hasAttemptedSave = React.useContext(SaveAttemptContext)
  const actionErrors = errors.actions?.[actionIndex] as any
  const allowedEntityTypes = selectedService?.allowed_targets
    .find((t) => Array.isArray(t.Type) && t.Type.includes("entity"))
    ?.EntityTypes

  const addOne = () => {
    updateAction(actionIndex, { ...action, targets: [...action.targets, { entity_id: "" }] })
  }

  const removeAt = (targetIndex: number) => {
    const newTargets = action.targets.filter((_, i) => i !== targetIndex)
    updateAction(actionIndex, { ...action, targets: newTargets })
  }

  return (
    <div className="space-y-2">
      <EnsureFirstTarget action={action} index={actionIndex} ensuresEnabled={true} updateAction={updateAction} />
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium text-gray-700">Targets</h4>
        <div className="flex items-center gap-2">
          <Button size="sm" variant="outline" onClick={addOne}>Add Target</Button>
        </div>
      </div>

      {action.targets.map((target, targetIndex) => {
        const targetErrors = actionErrors?.targets?.[targetIndex]
        return (
          <div key={targetIndex} className="space-y-1">
            <div className="flex items-center space-x-2">
              <EntitySelector
                value={target.entity_id!}
                allowedEntityTypes={allowedEntityTypes}
                onChange={(val) => {
                  const newTargets = [...action.targets]
                  newTargets[targetIndex] = { entity_id: val }
                  updateAction(actionIndex, { ...action, targets: newTargets })
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
            {hasAttemptedSave && targetErrors?.entity_id && (
              <p className="text-sm text-destructive ml-1">{targetErrors.entity_id.message}</p>
            )}
          </div>
        )
      })}
    </div>
  )
}
