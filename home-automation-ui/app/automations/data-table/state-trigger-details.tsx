"use client"

import { StateTrigger } from "@/types/automation"

export function StateTriggerDetails({ trigger }: { trigger: StateTrigger }) {
  return (
    <div className="flex flex-col space-y-0.5">
      <p>
        <span className="font-medium">Entity:</span> {trigger.entity_id}
      </p>
      {trigger.attribute && (
        <p>
          <span className="font-medium">Attribute:</span> {trigger.attribute}
        </p>
      )}
      {trigger.from !== undefined && (
        <p>
          <span className="font-medium">From:</span> {JSON.stringify(trigger.from)}
        </p>
      )}
      {trigger.to !== undefined && (
        <p>
          <span className="font-medium">To:</span> {JSON.stringify(trigger.to)}
        </p>
      )}
    </div>
  )
}
