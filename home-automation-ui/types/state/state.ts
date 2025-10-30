export interface Context {
  id: string
  parent_id?: string
}

export interface State {
  entity_id: string
  state: any
  attributes: Record<string, any>
  last_changed: string // ISO string
  last_updated: string // ISO string
  context?: Context
}
