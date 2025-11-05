import { z } from "zod";

// Target schema
const TargetSchema = z.object({
  entity_id: z.string().min(1, "Entity ID is required"),
  external_id: z.string().optional(),
});

// State trigger schema
const StateTriggerSchema = z.object({
  entity_id: z.string().min(1, "Entity ID is required"),
  attribute: z.string().optional(),
  from: z.any().optional(),
  to: z.any().optional(),
});

// Event trigger schema
const EventTriggerSchema = z.object({
  event_type: z.string().min(1, "Event type is required"),
});

// Base trigger schema (discriminated union)
const BaseTriggerSchema = z.discriminatedUnion("type", [
  z.object({
    type: z.literal("state"),
    data: StateTriggerSchema,
  }),
  z.object({
    type: z.literal("event"),
    data: EventTriggerSchema,
  }),
]);

// Helper function to check if a string contains template syntax
function containsTemplateSyntax(value: any): boolean {
  if (typeof value !== "string") return false;
  return value.includes("{{");
}

// Helper function to get API base URL (works in both client and server)
function getApiBaseUrl(): string {
  // In browser, use relative URL
  if (typeof window !== "undefined") {
    return "";
  }
  // On server, use absolute URL (assuming Next.js API routes proxy to engine)
  // For client-side validation, this will use relative URLs
  return typeof process !== "undefined" && process.env.NEXT_PUBLIC_API_URL
    ? process.env.NEXT_PUBLIC_API_URL
    : "";
}

// Helper function to validate template syntax via API
async function validateTemplateParam(
  service: string,
  paramName: string,
  value: string
): Promise<{ valid: boolean; error?: string }> {
  try {
    const baseUrl = getApiBaseUrl();
    const apiUrl = `${baseUrl}/api/services/validate-param`;
    
    const res = await fetch(apiUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        service,
        param_name: paramName,
        value,
      }),
    });

    if (!res.ok) {
      const data = await res.json();
      return {
        valid: false,
        error: data.error || "Template validation failed",
      };
    }

    return { valid: true };
  } catch (error: any) {
    return {
      valid: false,
      error: error.message || "Failed to validate template",
    };
  }
}

// Action schema with async validation
const ActionSchema = z
  .object({
    service: z.string().min(1, "Service is required"),
    targets: z.array(TargetSchema).min(0),
    params: z.record(z.string(), z.any()).optional(),
    blocking: z.boolean().optional(),
  })
  .superRefine(async (action, ctx) => {
    // Skip validation if service is not set
    if (!action.service) {
      return;
    }

    try {
      // Fetch service specs
      const baseUrl = getApiBaseUrl();
      const servicesUrl = `${baseUrl}/api/services`;
      const servicesRes = await fetch(servicesUrl);
      if (!servicesRes.ok) {
        ctx.addIssue({
          code: "custom",
          message: "Failed to fetch service specifications",
          path: ["service"],
        });
        return;
      }

      const servicesData = await servicesRes.json();
      const serviceSpec = servicesData.services?.find(
        (s: any) => s.name === action.service
      );

      if (!serviceSpec) {
        ctx.addIssue({
          code: "custom",
          message: `Service "${action.service}" not found`,
          path: ["service"],
        });
        return;
      }

      const requiredParams = serviceSpec.required_params || {};
      const actionParams = action.params || {};

      // Validate required params are present
      for (const [paramName] of Object.entries(requiredParams)) {
        if (!(paramName in actionParams)) {
          ctx.addIssue({
            code: "custom",
            message: `Required parameter "${paramName}" is missing`,
            path: ["params", paramName],
          });
        }
      }

      // Validate template syntax for all params
      for (const [paramName, paramValue] of Object.entries(actionParams)) {
        if (containsTemplateSyntax(paramValue)) {
          const validation = await validateTemplateParam(
            action.service,
            paramName,
            String(paramValue)
          );

          if (!validation.valid) {
            ctx.addIssue({
              code: "custom",
              message: validation.error || "Invalid template syntax",
              path: ["params", paramName],
            });
          }
        }
      }
    } catch (error: any) {
      ctx.addIssue({
        code: "custom",
        message: `Validation error: ${error.message || "Unknown error"}`,
        path: ["service"],
      });
    }
  });

// Main automation schema
export const CreateAutomationSchema = z.object({
  alias: z.string().min(1, "Alias is required"),
  description: z.string().optional(),
  triggers: z.array(BaseTriggerSchema).min(1, "At least one trigger is required"),
  actions: z.array(ActionSchema).min(1, "At least one action is required"),
  conditions: z.array(z.any()).optional(),
  enabled: z.boolean().optional(),
});
