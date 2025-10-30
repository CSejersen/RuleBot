export interface Automation {
  id?: number
  alias: string
  description: string
  triggers: BaseTrigger[]
  conditions: any[]
  actions: Action[]
  enabled: boolean
  last_triggered: string | null
  created_at?: string
  updated_at?: string
}

// ------------------- Triggers -------------------

export type TriggerType = 'state' | 'event';

export type BaseTrigger = {
  type: TriggerType;
  data: StateTrigger | EventTrigger;
};

export type StateTrigger = {
  entity_id: string;
  attribute?: string;
  from?: any;
  to?: any;
};

export type EventTrigger = {
  event_type: string;
};

// ------------------- Conditions -------------------
// Keep as any[] for now, we can expand later
export type Condition = any;

// ------------------- Actions -------------------

export type Action = {
  service: string;
  targets: Target[];
  params?: Record<string, any>;
  blocking?: boolean;
};

export type Target = {
  entity_id: string;
};
