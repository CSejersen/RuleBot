export interface Device {
  id: string;
  integration_id: number;
  type: string;
  name: string;
  metadata: Record<string, any>;
  enabled: boolean;
  available: boolean;
  created_at: string;
  updated_at: string;
}
