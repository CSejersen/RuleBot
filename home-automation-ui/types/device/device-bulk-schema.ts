import { z } from "zod";

export const DeviceBulkActionSchema = z.object({
  action: z.enum(["enable", "disable", "delete"]),
  ids: z.array(z.string().min(1, "id required")).min(1, "at least one id required"),
});

export type DeviceBulkAction = z.infer<typeof DeviceBulkActionSchema>;
