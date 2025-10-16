"use client"

import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select"
import { Rule } from "@/app/api/rules/types/rule"
import { ServiceData, ServicesResponse } from "@/app/api/services/route"


interface ActionSectionProps {
  newRule: Rule
  setNewRule: (rule: Rule) => void
}

export function ActionSection({ newRule, setNewRule }: ActionSectionProps) {
  const [services, setServices] = useState<ServiceData[]>([])

  // Fetch services on mount
  useEffect(() => {
    fetch("/api/services")
      .then((res) => res.json())
      .then((data: ServicesResponse) => setServices(data.services))
      .catch(console.error)
  }, [])

  const addAction = () => {
    setNewRule({
      ...newRule,
      action: [...newRule.action, { service: "", params: {} }],
    })
  }

  const removeAction = (index: number) => {
    const updated = [...newRule.action]
    updated.splice(index, 1)
    setNewRule({ ...newRule, action: updated })
  }

  return (
    <div className="border rounded-xl p-6 space-y-5">
      <div className="flex justify-between items-center">
        <h3 className="text-base font-semibold">Actions</h3>
        <Button variant="outline" size="sm" onClick={addAction}>
          + Add Action
        </Button>
      </div>

      {newRule.action.length === 0 && (
        <p className="text-sm text-muted-foreground">No actions added.</p>
      )}

      {newRule.action.map((act, index) => {
        const selectedService = services.find((s) => s.name === act.service)

        return (
          <div key={index} className="border rounded-md p-4 space-y-4 bg-muted/30">
            {/* Service Dropdown */}
            <div>
              <label className="block text-xs font-medium mb-1">Service</label>
              <Select
                value={act.service || ""}
                onValueChange={(val) => {
                  const updated = [...newRule.action]
                  updated[index].service = val
                  // reset params to default empty values
                  updated[index].params = {}
                  const service = services.find((s) => s.name === val)
                  if (service) {
                    Object.keys(service.required_params).forEach((param) => {
                      updated[index].params![param] = ""
                    })
                  }
                  setNewRule({ ...newRule, action: updated })
                }}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select Service" />
                </SelectTrigger>
                <SelectContent>
                  {services.map((s) => (
                    <SelectItem key={s.name} value={s.name}>
                      {s.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Target */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-medium mb-1">Target Type</label>
                <Input
                  value={act.target?.type || ""}
                  onChange={(e) => {
                    const updated = [...newRule.action]
                    updated[index].target = {
                      ...(updated[index].target || {}),
                      type: e.target.value || undefined,
                      id: updated[index].target?.id || "",
                    }
                    setNewRule({ ...newRule, action: updated })
                  }}
                  disabled={!selectedService?.requires_target_type}
                  placeholder={selectedService?.requires_target_type ? "" : "Not required"}
                />
              </div>
              <div>
                <label className="block text-xs font-medium mb-1">Target ID</label>
                <Input
                  value={act.target?.id || ""}
                  onChange={(e) => {
                    const updated = [...newRule.action]
                    updated[index].target = {
                      ...(updated[index].target || {}),
                      id: e.target.value,
                      type: updated[index].target?.type,
                    }
                    setNewRule({ ...newRule, action: updated })
                  }}
                  disabled={!selectedService?.requires_target_id}
                  placeholder={selectedService?.requires_target_id ? "" : "Not required"}
                />
              </div>
            </div>

            {/* Params */}
            {selectedService && Object.entries(selectedService.required_params).length > 0 && (
              <div>
                <span className="block text-xs font-medium mb-2">Params</span>
                {Object.entries(selectedService.required_params).map(([paramName, paramInfo]) => (
                  <div key={paramName} className="mb-2">
                    <label className="block text-xs font-medium">{paramName}</label>
                    <Input
                      placeholder={paramInfo.Description}
                      value={act.params?.[paramName] || ""}
                      onChange={(e) => {
                        const updated = [...newRule.action]
                        updated[index].params![paramName] = e.target.value
                        setNewRule({ ...newRule, action: updated })
                      }}
                    />
                  </div>
                ))}
              </div>
            )}

            {/* Remove Action */}
            <div className="flex justify-end">
              <Button variant="ghost" size="sm" onClick={() => removeAction(index)}>
                ✕ Remove Action
              </Button>
            </div>
          </div>
        )
      })}
    </div>
  )
}
