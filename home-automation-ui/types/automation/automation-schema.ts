import { z } from "zod";

export const CreateAutomationSchema = z.object({
  alias: z.string().min(1),
  description: z.string().optional(),
  triggers: z.array(z.any()).min(1), // Could be refined with your type
  actions: z.array(z.any()).min(1),  // Could be refined with your type
  conditions: z.array(z.any()).optional(),
  enabled: z.boolean().optional(),
});
