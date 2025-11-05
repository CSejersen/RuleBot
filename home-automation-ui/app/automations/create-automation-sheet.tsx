"use client"

import React from "react"
import { useForm, FormProvider } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { useEffect, useState } from "react"
import { Automation, BaseTrigger, Condition, Action } from "@/types/automation"
import { CreateAutomationSchema } from "@/types/automation/automation-schema"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription } from "@/components/ui/sheet"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import { TriggerSection } from "./form-sections/trigger"
import { ActionSection } from "./form-sections/action"
import { ChevronLeft, ChevronRight } from "lucide-react"
import { FormItem, FormLabel, FormControl, FormDescription, FormMessage } from "@/components/ui/form"

import { sendSocketMessage } from "@/lib/engine-socket"

// Context to track if save has been attempted
export const SaveAttemptContext = React.createContext<boolean>(false)

const STEPS = [
  { id: 0, label: "Details" },
  { id: 1, label: "Triggers" },
  { id: 2, label: "Actions" },
]

interface CreateAutomationSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onAutomationSaved: (automation: Automation) => void
  editingAutomation?: Automation | null
}

export const CreateAutomationSheet = /* @__PURE__ */ React.memo(function CreateAutomationSheet({
  open,
  onOpenChange,
  onAutomationSaved,
  editingAutomation = null,
}: CreateAutomationSheetProps) {
  const [currentStep, setCurrentStep] = useState(0)
  const [hasAttemptedSave, setHasAttemptedSave] = useState(false)
  const emptyAutomation: Automation = {
    alias: "",
    description: "",
    triggers: [],
    conditions: [],
    actions: [],
    enabled: true,
    last_triggered: null,
  }
  const methods = useForm<Automation>({
    resolver: zodResolver(CreateAutomationSchema),
    defaultValues: emptyAutomation,
    mode: "onSubmit",
  })


  useEffect(() => {
    if (!open) return
    if (editingAutomation) {
      methods.reset(editingAutomation)
    } else {
      methods.reset({
        ...emptyAutomation,
        triggers: [],
        actions: [],
      })
    }
    setCurrentStep(0)
    setHasAttemptedSave(false)
  }, [editingAutomation, open])

  const handleSave = async () => {
    // Mark that save has been attempted
    setHasAttemptedSave(true)
    // Validate form before submission
    const isValid = await methods.trigger()
    if (!isValid) {
      // Find first step with errors and navigate there
      const errors = methods.formState.errors
      if (errors.alias || errors.description) {
        setCurrentStep(0)
      } else if (errors.triggers) {
        setCurrentStep(1)
      } else if (errors.actions) {
        setCurrentStep(2)
      }
      return
    }

    const method = editingAutomation ? "PUT" : "POST"
    const url = editingAutomation
      ? `/api/automations/${editingAutomation.id}`
      : "/api/automations"

    const values = methods.getValues()
    const res = await fetch(url, {
      method,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(values),
    })

    const data = await res.json()
    if (!res.ok) {
      // Parse and display validation errors from API
      if (data.details?.fieldErrors) {
        // Map Zod field errors to form errors
        Object.entries(data.details.fieldErrors).forEach(([path, errorMessages]) => {
          const messages = Array.isArray(errorMessages) ? errorMessages : [String(errorMessages)]
          if (messages.length > 0) {
            // Convert path like "triggers.0.data.entity_id" to nested path array
            const fieldPath = path.split(".") as any[]
            methods.setError(fieldPath, {
              type: "server",
              message: messages[0],
            })
          }
        })
        // Show form-level errors if any
        if (data.details?.formErrors?.length > 0) {
          alert(data.details.formErrors.join(", "))
        }
      } else {
        // Fallback for non-Zod errors
        alert(data.error || "Failed to save automation")
      }
      return
    }

    onAutomationSaved(data.automation)
    onOpenChange(false)
    methods.reset(emptyAutomation)
    sendSocketMessage({ type: "reload_automations" })
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-1/3 flex flex-col">
        <SheetHeader>
          <SheetTitle>{editingAutomation ? "Edit Automation" : "Create New Automation"}</SheetTitle>
          <SheetDescription>
            {editingAutomation
              ? "Update the details for this automation below."
              : "Fill in the details for your new automation below."}
          </SheetDescription>
        </SheetHeader>

        {/* Step Progress Indicator */}
        <div className="px-6 pt-6 pb-4">
          <div className="flex items-center">
            {STEPS.map((step, index) => (
              <React.Fragment key={step.id}>
                <div className="flex flex-col items-center flex-1 relative">
                  <div className="relative w-full flex items-center">
                    <div
                      className={`w-10 h-10 rounded-full flex items-center justify-center border-2 font-semibold text-sm transition-colors mx-auto relative z-10 ${
                        currentStep === step.id
                          ? "border-primary bg-primary text-primary-foreground"
                          : currentStep > step.id
                          ? "border-primary bg-primary text-primary-foreground"
                          : "border-muted bg-background text-muted-foreground"
                      }`}
                    >
                      {currentStep > step.id ? (
                        <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                          <path
                            fillRule="evenodd"
                            d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                            clipRule="evenodd"
                          />
                        </svg>
                      ) : (
                        step.id + 1
                      )}
                    </div>
                    {index < STEPS.length - 1 && (
                      <div
                        className={`absolute h-0.5 transition-colors ${
                          currentStep > step.id ? "bg-primary" : "bg-muted"
                        }`}
                        style={{
                          left: "calc(50% + 20px)", // Start from circle center (50% container + half circle width)
                          width: "100%", // Extend to center of next container
                          top: "50%",
                          transform: "translateY(-50%)",
                        }}
                      />
                    )}
                  </div>
                  <div className="mt-2 text-center">
                    <div
                      className={`text-xs font-medium ${
                        currentStep === step.id ? "text-foreground" : "text-muted-foreground"
                      }`}
                    >
                      {step.label}
                    </div>
                  </div>
                </div>
              </React.Fragment>
            ))}
          </div>
        </div>

        <FormProvider {...methods}>
          <SaveAttemptContext.Provider value={hasAttemptedSave}>
            <div className="overflow-y-auto flex-1 px-6 mt-4">
              {currentStep === 0 && (
                <div className="space-y-6">
                  <FormItem>
                    <FormLabel>Alias</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="My Automation"
                        {...methods.register("alias")}
                      />
                    </FormControl>
                    {hasAttemptedSave && methods.formState.errors.alias && (
                      <p className="text-sm text-destructive mt-1">{methods.formState.errors.alias.message}</p>
                    )}
                  </FormItem>

                  <FormItem>
                    <FormLabel>Description</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder="What does this automation do?"
                        {...methods.register("description")}
                        className="resize-none"
                        rows={3}
                      />
                    </FormControl>
                    {hasAttemptedSave && methods.formState.errors.description && (
                      <p className="text-sm text-destructive mt-1">{methods.formState.errors.description.message}</p>
                    )}
                  </FormItem>

                  <div className="flex items-center space-x-3">
                    <Switch
                      checked={!!methods.watch("enabled")}
                      onCheckedChange={(checked) => methods.setValue("enabled", checked, { shouldDirty: true })}
                    />
                    <span className="text-sm font-medium">Enabled</span>
                  </div>
                </div>
              )}

              {currentStep === 1 && (
                <TriggerSection />
              )}

              {currentStep === 2 && (
                <ActionSection />
              )}
            </div>
          </SaveAttemptContext.Provider>
        </FormProvider>

        {/* Navigation Footer */}
        <div className="border-t mt-6 px-6 py-4 flex justify-between items-center bg-background space-x-2">
          <Button
            variant="outline"
            onClick={() => setCurrentStep((prev) => Math.max(0, prev - 1))}
            disabled={currentStep === 0}
          >
            <ChevronLeft className="w-4 h-4 mr-2" />
            Back
          </Button>
          <div className="flex space-x-2">
            {currentStep === STEPS.length - 1 ? (
              <Button onClick={handleSave}>
                {editingAutomation ? "Update Automation" : "Save Automation"}
              </Button>
            ) : (
              <Button
                onClick={async () => {
                  // Mark that validation has been attempted so errors will show
                  setHasAttemptedSave(true)
                  // Validate current step before proceeding
                  let isValid = true
                  if (currentStep === 0) {
                    // Trigger validation and mark fields as touched so errors show
                    isValid = await methods.trigger("alias")
                    await methods.trigger("description")
                    if (!isValid) {
                      // Mark fields as touched to ensure errors are displayed
                      methods.setFocus("alias")
                    }
                  } else if (currentStep === 1) {
                    isValid = await methods.trigger("triggers")
                  } else if (currentStep === 2) {
                    isValid = await methods.trigger("actions")
                  }
                  if (isValid) {
                    setCurrentStep((prev) => Math.min(STEPS.length - 1, prev + 1))
                  }
                }}
              >
                Next
                <ChevronRight className="w-4 h-4 ml-2" />
              </Button>
            )}
          </div>
        </div>
      </SheetContent>
    </Sheet>
  )
})
