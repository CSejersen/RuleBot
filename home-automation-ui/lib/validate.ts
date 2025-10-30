import { ZodSchema } from 'zod';

/**
 * Validates request data (body, query, etc.) against a Zod schema. Returns success/data or formatted error.
 */
export function validateRequest<T>(schema: ZodSchema<T>, data: unknown):
  | { success: true; data: T }
  | { success: false; error: { error: string; details: any } } {
  const result = schema.safeParse(data);
  if (result.success) {
    return { success: true, data: result.data };
  }
  return {
    success: false,
    error: {
      error: 'Invalid request',
      details: result.error.flatten(),
    },
  };
}
