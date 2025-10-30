export interface ServiceParam {
    dataType: string;
    description: string;
  }
  
  export type TargetType = "entity"; // Currently only "entity" is supported
  
  export type EntityType = "light" | "scene" | "speaker" | "button" | "unknown";
  
  export interface TargetSpec {
    type: TargetType[];
    entityTypes?: EntityType[];
  }
  
  export interface ServiceSpec {
    name: string;
    requiredParams: Record<string, ServiceParam>;
    allowedTargets: TargetSpec;
  }
  
  export interface ServicesResponse {
    services: ServiceSpec[];
    error?: string;
  }