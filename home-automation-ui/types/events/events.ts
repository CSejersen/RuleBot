export type EventType = 'state_changed' | 'call_service' | 'time_changed';

export type State = {
  entity_id: string;
  state: any;
  attributes: Record<string, any>;
  last_changed: string;
  last_updated: string;
  context?: { id: string; parent_id?: string; created_at?: string };
};

export type StateChangedData = {
  entity_id: string;
  old_state: State | null;
  new_state: State | null;
};

export type CallServiceData = {
  domain: string;
  service: string;
  service_data: Record<string, any>;
  entity_id?: string;
};

export type Event = {
  id: string;
  type: EventType;
  data: StateChangedData | CallServiceData | Record<string, any>;
  context_id: string | null;
  time_fired: string;
  context?: {
    id: string;
    parent_id?: string;
    created_at?: string;
  };
  created_at?: string;
  updated_at?: string;
};
