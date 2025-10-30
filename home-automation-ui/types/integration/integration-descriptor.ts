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
