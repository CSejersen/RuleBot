"use client"

import React from "react"
import { useForm, FormProvider } from "react-hook-form"
import { useEffect, useState } from "react"
import { Automation, BaseTrigger, Condition, Action } from "@/types/automation"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription } from "@/components/ui/sheet"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import { TriggerSection } from "./form-sections/trigger"
import { ActionSection } from "./form-sections/action"
import { ChevronLeft, ChevronRight } from "lucide-react"

import { engineWSsendMessage } from "@/lib/engine-socket"

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
    defaultValues: emptyAutomation,
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
  }, [editingAutomation, open])

  const handleSave = async () => {
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
    if (!res.ok) return alert(data.error || "Failed to save automation")

    onAutomationSaved(data.automation)
    onOpenChange(false)
    methods.reset(emptyAutomation)
    engineWSsendMessage({ type: "reload_automations" })
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

        <div className="overflow-y-auto flex-1 px-6 mt-4">
          {currentStep === 0 && (
            <div className="space-y-6">
              <div>
                <h3 className="text-base font-semibold mb-2">Alias</h3>
                <Input
                  placeholder="My Automation"
                  value={methods.watch("alias")}
                  onChange={(e) => methods.setValue("alias", e.target.value, { shouldDirty: true })}
                />
              </div>

              <div>
                <h3 className="text-base font-semibold mb-2">Description</h3>
                <Textarea
                  placeholder="What does this automation do?"
                  value={methods.watch("description")}
                  onChange={(e) => methods.setValue("description", e.target.value, { shouldDirty: true })}
                  className="resize-none"
                  rows={3}
                />
              </div>

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
            <FormProvider {...methods}>
              <TriggerSection />
            </FormProvider>
          )}

          {currentStep === 2 && (
            <FormProvider {...methods}>
              <ActionSection />
            </FormProvider>
          )}
        </div>

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
                onClick={() => setCurrentStep((prev) => Math.min(STEPS.length - 1, prev + 1))}
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
