"use client"

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog"
import { Rule } from "@/app/api/rules/types/rule"
import {
  TriggerDetails,
  ConditionDetails,
  ActionDetails,
} from "./rule-detail-sections"

interface RuleDetailsDialogProps {
  rule: Rule | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function RuleDetailsDialog({ rule, open, onOpenChange }: RuleDetailsDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle className="text-lg font-semibold">Rule Details</DialogTitle>
          <DialogDescription className="text-sm text-muted-foreground">
            {rule ? rule.alias : "Select a rule to see details."}
          </DialogDescription>
        </DialogHeader>

        {rule ? (
          <div className="space-y-4 text-sm">
            <p>
              <span className="font-medium">Last Triggered:</span>{" "}
              {rule.lastTriggered
                ? new Date(rule.lastTriggered).toLocaleString("en-GB", {
                  day: "2-digit",
                  month: "short",
                  year: "numeric",
                  hour: "2-digit",
                  minute: "2-digit",
                  second: "2-digit",
                  hour12: false,
                })
                : "—"}
            </p>
            <p>
              <span className="font-medium">Active:</span> {rule.active ? "Yes" : "No"}
            </p>

            <TriggerDetails trigger={rule.trigger} />

            {rule.condition?.length > 0 && (
              <ConditionDetails conditions={rule.condition} />
            )}

            {rule.action?.length > 0 && <ActionDetails actions={rule.action} />}
          </div>
        ) : (
          <p className="text-sm">No rule selected.</p>
        )}
      </DialogContent>
    </Dialog>
  )
}
