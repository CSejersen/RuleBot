"use client"

import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { StateTriggerDetails } from "./state-trigger-details"
import { EventTriggerDetails } from "./event-trigger-details"
import { Action, BaseTrigger, Condition, EventTrigger, StateTrigger } from "@/types/automation"

export function TriggerDetails({ triggers }: { triggers: BaseTrigger[] }) {
  if (!triggers || triggers.length === 0) return null

  return (
    <div className="space-y-1">
      <p className="font-semibold mb-1">Triggers:</p>
      <div className="flex flex-col space-y-1 bg-muted p-3 rounded-md text-sm">
        {triggers.map((trigger, index) => (
          <div key={index} className="flex flex-col space-y-1">
            <p>
              <span className="font-medium">Type:</span> {trigger.type}
            </p>
            {trigger.type === "state" && (trigger.data as StateTrigger) && (
              <StateTriggerDetails trigger={trigger.data as StateTrigger} />
            )}
            {trigger.type === "event" && (trigger.data as EventTrigger) && (
              <EventTriggerDetails trigger={trigger.data as EventTrigger} />
            )}
          </div>
        ))}
      </div>
    </div>
  )
}

export function ConditionDetails({ conditions }: { conditions: Condition[] }) {
  return (
    <div className="space-y-1">
      <p className="font-semibold mb-1">Conditions:</p>
      <div className="flex flex-col space-y-1 bg-muted p-3 rounded-md text-sm">
        {conditions.map((cond, index) => (
          <div key={index} className="flex flex-col space-y-0.5">
            <p>
              <span className="font-medium">Entity:</span> {cond.entity}
            </p>
            <p>
              <span className="font-medium">Field:</span> {cond.field}
            </p>
            {cond.equals !== undefined && (
              <p>
                <span className="font-medium">Equals:</span> {JSON.stringify(cond.equals)}
              </p>
            )}
            {cond.notEquals !== undefined && (
              <p>
                <span className="font-medium">Not Equals:</span> {JSON.stringify(cond.notEquals)}
              </p>
            )}
            {cond.gt !== undefined && (
              <p>
                <span className="font-medium">Greater Than:</span> {cond.gt}
              </p>
            )}
            {cond.lt !== undefined && (
              <p>
                <span className="font-medium">Less Than:</span> {cond.lt}
              </p>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}

export function ActionDetails({ actions }: { actions: Action[] }) {
  if (!actions || actions.length === 0) return null

  return (
    <div className="space-y-1">
      <p className="font-semibold mb-1">Actions:</p>
      <div className="flex flex-col space-y-1 bg-muted p-3 rounded-md text-sm">
        <Accordion type="multiple" className="flex flex-col space-y-1">
          {actions.map((action, index) => (
            <AccordionItem key={index} value={`action-${index}`}>
              <AccordionTrigger className="p-2 text-sm font-medium">{action.service}</AccordionTrigger>
              <AccordionContent className="pl-2 pb-2 pt-1 text-sm space-y-1">
                {action.targets?.length > 0 && (
                  <div>
                    <span className="font-medium">Targets:</span>
                    <ul className="list-disc list-inside ml-2">
                      {action.targets.map((target, tIndex) => (
                        <li key={tIndex}>{target.entity_id}</li>
                      ))}
                    </ul>
                  </div>
                )}
                {action.params && (
                  <div>
                    <span className="font-medium">Params:</span>
                    <pre className="bg-muted p-2 rounded-md overflow-auto text-sm">
                      {JSON.stringify(action.params, null, 2)}
                    </pre>
                  </div>
                )}
                <p>
                  <span className="font-medium">Blocking:</span> {action.blocking ? "Yes" : "No"}
                </p>
              </AccordionContent>
            </AccordionItem>
          ))}
        </Accordion>
      </div>
    </div>
  )
}
