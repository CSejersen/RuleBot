"use client"

import { EntitySelector } from "@/components/selectors/entity-selector"
import { TargetSpec } from "@/types/service"

interface TargetSelectorProps {
  spec: TargetSpec
  value: any
  onChange: (val: any) => void
}

export function TargetSelector({ spec, value, onChange }: TargetSelectorProps) {
  if (spec.type.includes("entity")) {
    return (
      <EntitySelector
        allowedEntityTypes={spec.entityTypes}
        value={value?.entity_id || ""}
        onChange={(entityId) => onChange({ entity_id: entityId })}
      />
    )
  }

  return <div className="text-sm text-muted-foreground">Unsupported target type</div>
}
