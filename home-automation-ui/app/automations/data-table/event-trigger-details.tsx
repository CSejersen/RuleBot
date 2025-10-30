"use client"

import { EventTrigger } from "@/types/automation"

export function EventTriggerDetails({ trigger }: { trigger: EventTrigger }) {
  return (
    <div className="flex flex-col space-y-0.5">
      <p>
        <span className="font-medium">Event Type:</span> {trigger.event_type}
      </p>
    </div>
  )
}
