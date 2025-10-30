import { z } from 'zod';

export const IntegrationDiscoverParamsSchema = z.object({
  name: z.string().min(1),
});
