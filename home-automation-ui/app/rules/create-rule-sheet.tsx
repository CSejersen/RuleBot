"use client"

import { useState } from "react"
import { Rule } from "@/app/api/rules/types/rule"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription } from "@/components/ui/sheet"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Switch } from "@/components/ui/switch"
import { TriggerSection } from "./form-sections/trigger"
import { ConditionSection } from "./form-sections/condition"
import { ActionSection } from "./form-sections/action"

interface CreateRuleSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onRuleCreated: (rule: Rule) => void
}

export function CreateRuleSheet({ open, onOpenChange, onRuleCreated }: CreateRuleSheetProps) {
  const [newRule, setNewRule] = useState<Rule>({
    alias: "",
    trigger: { event: "", entityName: "", stateChange: "" },
    condition: [],
    action: [],
    active: true,
    lastTriggered: null,
  })

  const handleSave = async () => {
    const res = await fetch("/api/rules", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(newRule),
    })
    const data = await res.json()
    if (!res.ok) return alert(data.error || "Failed to create rule")
    onRuleCreated(data.rule)
    onOpenChange(false)
    setNewRule({
      alias: "",
      trigger: { event: "", entityName: "", stateChange: "" },
      condition: [],
      action: [],
      active: true,
      lastTriggered: null,
    })
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-1/3 flex flex-col">
        <SheetHeader>
          <SheetTitle>Create New Rule</SheetTitle>
          <SheetDescription>Fill in the details for your new rule below.</SheetDescription>
        </SheetHeader>

        <div className="overflow-y-auto flex-1 px-6 mt-6 space-y-8">
          <div>
            <h3 className="text-base font-semibold mb-2">Rule Name</h3>
            <Input
              value={newRule.alias}
              onChange={(e) => setNewRule({ ...newRule, alias: e.target.value })}
            />
          </div>

          <TriggerSection newRule={newRule} setNewRule={setNewRule} />
          <ConditionSection newRule={newRule} setNewRule={setNewRule} />
          <ActionSection newRule={newRule} setNewRule={setNewRule} />

          <div className="flex items-center space-x-3">
            <Switch
              checked={newRule.active}
              onCheckedChange={(checked) => setNewRule({ ...newRule, active: checked })}
            />
            <span className="text-sm font-medium">Active</span>
          </div>
        </div>

        <div className="border-t mt-6 px-6 py-4 flex justify-end bg-background">
          <Button onClick={handleSave}>Save Rule</Button>
        </div>
      </SheetContent>
    </Sheet>
  )
}
