export interface Entity {
  external_id: string
  device_id?: string
  entity_id: string
  type: string
  name: string
  available: boolean
  enabled: boolean
}
