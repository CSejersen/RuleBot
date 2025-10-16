"use client"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Rule } from "../types/rule"

interface ConditionSectionProps {
  newRule: Rule
  setNewRule: (rule: Rule) => void
}

export function ConditionSection({ newRule, setNewRule }: ConditionSectionProps) {
  const addCondition = () => {
    setNewRule({
      ...newRule,
      condition: [...newRule.condition, { entity: "", field: "", equals: "" }],
    })
  }

  const removeCondition = (index: number) => {
    const updated = [...newRule.condition]
    updated.splice(index, 1)
    setNewRule({ ...newRule, condition: updated })
  }

  const updateCondition = (index: number, field: string, value: string) => {
    const updated = [...newRule.condition]
    updated[index][field as keyof typeof updated[number]] = value
    setNewRule({ ...newRule, condition: updated })
  }

  const updateOperator = (index: number, operator: string) => {
    const updated = [...newRule.condition]
    const cond = updated[index]
      // remove all operators
      ;["equals", "notEquals", "gt", "lt"].forEach((op) => delete cond[op as keyof typeof cond])
    cond[operator as keyof typeof cond] = ""
    setNewRule({ ...newRule, condition: updated })
  }

  return (
    <div className="border rounded-xl p-6 space-y-5">
      <div className="flex justify-between items-center">
        <h3 className="text-base font-semibold">Conditions (optional)</h3>
        <Button variant="outline" size="sm" onClick={addCondition}>
          + Add Condition
        </Button>
      </div>

      {newRule.condition.length === 0 && (
        <p className="text-sm text-muted-foreground">No conditions added.</p>
      )}

      {newRule.condition.map((cond, index) => {
        const activeOp =
          Object.keys(cond).find((k) => ["equals", "notEquals", "gt", "lt"].includes(k)) || "equals"

        return (
          <div key={index} className="border rounded-md p-4 space-y-4 bg-muted/30">
            {/* Entity */}
            <div>
              <label className="block text-xs font-medium mb-1">Entity</label>
              <Input
                value={cond.entity}
                onChange={(e) => updateCondition(index, "entity", e.target.value)}
              />
            </div>

            {/* Field */}
            <div>
              <label className="block text-xs font-medium mb-1">Field</label>
              <Input
                value={cond.field}
                onChange={(e) => updateCondition(index, "field", e.target.value)}
              />
            </div>

            {/* Operator + Value */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-medium mb-1">Operator</label>
                <select
                  className="w-full border rounded-md h-9 text-sm px-2"
                  value={activeOp}
                  onChange={(e) => updateOperator(index, e.target.value)}
                >
                  <option value="equals">Equals</option>
                  <option value="notEquals">Not Equals</option>
                  <option value="gt">Greater Than</option>
                  <option value="lt">Less Than</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-medium mb-1">Value</label>
                <Input
                  value={cond.equals ?? cond.notEquals ?? cond.gt ?? cond.lt ?? ""}
                  onChange={(e) => {
                    const op = activeOp
                    const updated = [...newRule.condition]
                    updated[index][op as keyof typeof cond] = e.target.value
                    setNewRule({ ...newRule, condition: updated })
                  }}
                />
              </div>
            </div>

            {/* Remove Button */}
            <div className="flex justify-end">
              <Button variant="ghost" size="sm" onClick={() => removeCondition(index)}>
                ✕ Remove
              </Button>
            </div>
          </div>
        )
      })}
    </div>
  )
}
