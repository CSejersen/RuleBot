export type Rule = {
  alias: string;
  trigger: Trigger;
  condition: Condition[];
  action: Action[];
  active: boolean;
  lastTriggered: string | null;
};

export type Trigger = {
  event: string;
  entityName?: string; // optional
  stateChange: string;
};

export type Condition = {
  entity: string; // Integration.Typ.Entity_name
  field: string;  // e.g. "brightness"

  // Comparison operators
  equals?: any;
  notEquals?: any;
  gt?: number;
  lt?: number;
};

export type Action = {
  service: string;
  target?: Target;
  params?: Record<string, any>;
  blocking?: boolean;
};

export type Target = {
  type?: string;
  id: string;
};
