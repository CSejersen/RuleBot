
"use client"

import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Condition, Action, Trigger } from "@/app/api/rules/types/rule"

export function TriggerDetails({ trigger }: { trigger: Trigger }) {
  return (
    <div className="space-y-1">
      <p className="font-semibold mb-1">Trigger:</p>
      <div className="flex flex-col space-y-1 bg-muted p-3 rounded-md text-sm">
        <p>
          <span className="font-medium">Event:</span> {trigger.event}
        </p>
        {trigger.entityName && (
          <p>
            <span className="font-medium">Entity:</span> {trigger.entityName}
          </p>
        )}
        <p>
          <span className="font-medium">State Change:</span> {trigger.stateChange}
        </p>
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
  return (
    <div className="space-y-1">
      <p className="font-semibold mb-1">Actions:</p>
      <div className="flex flex-col space-y-1 bg-muted p-3 rounded-md text-sm">
        <Accordion type="multiple" className="flex flex-col space-y-1">
          {actions.map((action, index) => (
            <AccordionItem key={index} value={`action-${index}`}>
              <AccordionTrigger className="p-2 text-sm font-medium">
                {action.service}
              </AccordionTrigger>
              <AccordionContent className="pl-2 pb-2 pt-1 text-sm space-y-1">
                {action.target && (
                  <p>
                    <span className="font-medium">Target:</span>{" "}
                    {action.target.type
                      ? `${action.target.type} / ${action.target.id}`
                      : action.target.id}
                  </p>
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
                  <span className="font-medium">Blocking:</span>{" "}
                  {action.blocking ? "Yes" : "No"}
                </p>
              </AccordionContent>
            </AccordionItem>
          ))}
        </Accordion>
      </div>
    </div>
  )
}
