"use client"

import { useEffect, useState } from "react"
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select"

interface TriggerSectionProps {
  newRule: any
  setNewRule: (rule: any) => void
}

interface EventType {
  type: string
  entities: string[]
  stateChanges: string[]
}

export function TriggerSection({ newRule, setNewRule }: TriggerSectionProps) {
  const [integrations, setIntegrations] = useState<string[]>([])
  const [eventTypes, setEventTypes] = useState<EventType[]>([])
  const [loadingEvents, setLoadingEvents] = useState(false)

  const eventParts = (newRule.trigger?.event || "").split(".")
  const source = eventParts[0] || ""
  const eventType = eventParts[1] || ""

  const currentEvent = eventTypes.find((et) => et.type === eventType)
  const entities = currentEvent?.entities || []
  const stateChanges = currentEvent?.stateChanges || []

  // Fetch integrations on mount
  useEffect(() => {
    fetch("/api/integrations")
      .then((res) => res.json())
      .then((data) => setIntegrations(data.integrations))
      .catch(console.error)
  }, [])

  // Fetch event types when source changes
  useEffect(() => {
    if (!source) {
      setEventTypes([])
      setNewRule({
        ...newRule,
        trigger: { ...newRule.trigger, event: "" },
      })
      return
    }

    setLoadingEvents(true)
    fetch(`/api/integrations/${source}/events`)
      .then((res) => res.json())
      .then((data) => {
        setEventTypes(data.events || [])
        // Reset the eventType part of the event to empty
        setNewRule({
          ...newRule,
          trigger: { ...newRule.trigger, event: `${source}.` },
        })
      })
      .finally(() => setLoadingEvents(false))
  }, [source])

  return (
    <div className="border rounded-xl p-6 space-y-5">
      <h3 className="text-base font-semibold mb-2">Trigger</h3>

      {/* Source + Event Type side by side */}
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between sm:gap-4">
        {/* Source */}
        <div className="flex-1">
          <label className="block text-sm font-medium mb-1">Source</label>
          <Select
            value={source || ""}
            onValueChange={(val) =>
              setNewRule({
                ...newRule,
                trigger: {
                  ...newRule.trigger,
                  event: `${val}.${eventType || ""}`,
                  entityName: undefined, // reset when source changes
                },
              })
            }
          >
            <SelectTrigger>
              <SelectValue placeholder="Select Source" />
            </SelectTrigger>
            <SelectContent>
              {integrations.map((i) => (
                <SelectItem key={i} value={i}>
                  {i}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Event Type */}
        <div className="flex-1">
          <label className="block text-sm font-medium mb-1">Event Type</label>
          <Select
            value={eventType || ""}
            onValueChange={(val) =>
              setNewRule({
                ...newRule,
                trigger: {
                  ...newRule.trigger,
                  event: `${source || ""}.${val}`,
                  entityName: undefined, // reset entity when event changes
                  stateChange: undefined, // reset stateChange when event changes
                },
              })
            }
            disabled={!source || loadingEvents}
          >
            <SelectTrigger className={`${!source ? "bg-gray-100" : ""}`}>
              <SelectValue
                placeholder={loadingEvents ? "Loading..." : "Select Event Type"}
              />
            </SelectTrigger>
            <SelectContent>
              {eventTypes.map((e) => (
                <SelectItem key={e.type} value={e.type}>
                  {e.type}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Entity Name */}
      <div>
        <label className="block text-sm font-medium mb-1">Entity Name</label>
        <Select
          value={newRule.trigger.entityName || ""}
          onValueChange={(val) =>
            setNewRule({
              ...newRule,
              trigger: { ...newRule.trigger, entityName: val },
            })
          }
          disabled={entities.length === 0}
        >
          <SelectTrigger className={`${entities.length === 0 ? "bg-gray-100" : ""}`}>
            <SelectValue
              placeholder={
                entities.length > 0
                  ? "Select Entity"
                  : "No entity selection required for this event type"
              }
            />
          </SelectTrigger>
          <SelectContent>
            {entities.map((name) => (
              <SelectItem key={name} value={name}>
                {name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* State Change */}
      <div>
        <label className="block text-sm font-medium mb-1">State Change</label>
        <Select
          value={newRule.trigger.stateChange || ""}
          onValueChange={(val) =>
            setNewRule({
              ...newRule,
              trigger: { ...newRule.trigger, stateChange: val },
            })
          }
          disabled={!eventType}
        >
          <SelectTrigger className={`${!eventType ? "bg-gray-100" : ""}`}>
            <SelectValue placeholder="Select State Change" />
          </SelectTrigger>
          <SelectContent>
            {stateChanges.map((sc) => (
              <SelectItem key={sc} value={sc}>
                {sc}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    </div>
  )
}
