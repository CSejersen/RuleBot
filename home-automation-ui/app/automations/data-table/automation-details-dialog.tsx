"use client"

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Automation } from "@/types/automation"
import { TriggerDetails, ConditionDetails, ActionDetails } from "./automation-details-sections"
import { Badge } from "@/components/ui/badge"

import { engineWSsendMessage } from "@/lib/engine-socket"
import React from "react"

interface AutomationDetailsDialogProps {
  automation: Automation | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onEdit: (automation: Automation) => void
  onDeleteOptimistic: (automation: Automation) => Promise<void>
}

function AutomationDetailsDialogInner({
  automation,
  open,
  onOpenChange,
  onEdit,
  onDeleteOptimistic,
}: AutomationDetailsDialogProps) {
  if (typeof window !== "undefined") console.count("Render: AutomationDetailsDialog")
  if (!automation) return null

  const handleDelete = async () => {
    if (typeof window !== "undefined") console.log("Dialog handleDelete", automation.id)
    if (!automation.id) {
      alert("Cannot delete an automation without an ID")
      return
    }

    if (!confirm(`Are you sure you want to delete the automation "${automation.alias}"?`)) return

    await onDeleteOptimistic(automation)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle className="text-lg font-semibold">{automation.alias || "Automation Details"}</DialogTitle>
          {automation.description && (
            <DialogDescription className="text-sm text-muted-foreground">
              {automation.description}
            </DialogDescription>
          )}
        </DialogHeader>

        <div className="space-y-4 text-sm max-h-[70vh] overflow-y-auto pr-1">
          <div className="flex items-center gap-2">
            <Badge variant={automation.enabled ? "secondary" : "outline"}>
              {automation.enabled ? "Enabled" : "Disabled"}
            </Badge>
            <span className="text-muted-foreground">
              Last triggered: {automation.last_triggered
                ? (() => {
                    const iso = automation.last_triggered.endsWith("Z") ? automation.last_triggered : automation.last_triggered + "Z"
                    const d = new Date(iso)
                    return isNaN(d.getTime()) ? "Invalid date" : d.toLocaleString("en-GB", { year: "numeric", month: "short", day: "2-digit", hour: "2-digit", minute: "2-digit", second: "2-digit", hour12: false })
                  })()
                : "—"}
            </span>
          </div>

          <div className="text-muted-foreground">ID: {automation.id ?? "—"}</div>

          <TriggerDetails triggers={automation.triggers ?? []} />

          {(automation.conditions?.length ?? 0) > 0 && (
            <ConditionDetails conditions={automation.conditions!} />
          )}

          {(automation.actions?.length ?? 0) > 0 && <ActionDetails actions={automation.actions!} />}
        </div>

        <div className="flex justify-between mt-6">
          {/* Delete button on the left */}
          <Button variant="destructive" size="sm" onClick={handleDelete}>
            Delete
          </Button>

          {/* Close/Edit buttons on the right */}
          <div className="space-x-2">
            <Button variant="outline" size="sm" onClick={() => onOpenChange(false)}>
              Close
            </Button>
            <Button size="sm" onClick={() => onEdit(automation)}>
              Edit
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

export const AutomationDetailsDialog = React.memo(AutomationDetailsDialogInner)
