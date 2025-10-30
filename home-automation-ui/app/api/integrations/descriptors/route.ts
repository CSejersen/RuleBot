"use server";

import { NextResponse } from "next/server";

export interface ConfigField {
  name: string;
  label: string;
  description?: string;
  placeholder?: string;
  type: "text" | "password" | "number" | "checkbox" | "select";
  required: boolean;
  default?: any;
  options?: string[];
}

export interface ConfigSchema {
  fields: ConfigField[];
}

export interface IntegrationDescriptor {
  name: string;
  display_name: string;
  description: string;
  version: string;
  capabilities: string[];
  config_schema: ConfigSchema;
}

export async function GET() {
  try {
    const res = await fetch("http://localhost:8080/api/integrations/descriptors", {
      headers: { "Content-Type": "application/json" },
    });

    if (!res.ok) {
      console.error("Engine responded with error:", res.status);
      return NextResponse.json({ error: "Failed to fetch integration" }, { status: res.status });
    }

    const data = await res.json(); // { descriptors: [...] }

    // Transform map<string, ConfigField> -> { fields: ConfigField[] }
    const descriptors: IntegrationDescriptor[] = (data.descriptors ?? []).map((d: any) => ({
      name: d.name,
      display_name: d.display_name,
      description: d.description,
      version: d.version,
      capabilities: d.capabilities,
      config_schema: {
        fields: Object.entries(d.config_schema || {}).map(([key, field]: [string, any]) => ({
          name: key,
          label: field.label,
          description: field.description,
          placeholder: field.placeholder,
          type: field.type,
          required: field.required,
          default: field.default,
          options: field.options,
        })),
      },
    }));

    return NextResponse.json(descriptors);
  } catch (error) {
    console.error("Error fetching integration:", error);
    return NextResponse.json({ error: "Internal server error" }, { status: 500 });
  }
}
