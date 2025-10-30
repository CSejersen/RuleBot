import { z } from 'zod';

export const IntegrationConfigSchema = z.object({
  integration_name: z.string().min(1),
  display_name: z.string().optional(),
  user_config: z.record(z.string(), z.unknown()),
  enabled: z.boolean().optional(),
});
