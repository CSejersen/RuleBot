"use client";

import { useEffect, useState } from "react";
import { useForm, SubmitHandler } from "react-hook-form";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Loader2 } from "lucide-react";
import { ConfigSchema } from "../api/integrations/descriptors/route";
import { IntegrationConfig } from "../api/integrations/configs/route";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { engineWSsendMessage } from "@/lib/engine-socket";

interface AddIntegrationDialogProps {
  open: boolean;
  onClose: () => void;
  descriptor: {
    name: string;
    display_name: string;
    config_schema: ConfigSchema;
  } | null;
  onCreated: (newConfig: IntegrationConfig) => void;
}

export function AddIntegrationDialog({ open, onClose, descriptor, onCreated }: AddIntegrationDialogProps) {
  const { register, handleSubmit, reset, setValue } = useForm<Record<string, any>>();
  const [loading, setLoading] = useState(false);

  // Reset form when dialog opens
  useEffect(() => {
    if (descriptor) {
      const defaults: Record<string, any> = {};
      descriptor.config_schema?.fields?.forEach((f) => {
        if (f.default !== undefined) {
          defaults[f.name] = f.default;
        } else if (f.type === "checkbox") {
          defaults[f.name] = false;
        }
      });
      reset(defaults);
    }
  }, [descriptor, reset]);

  const onSubmit: SubmitHandler<Record<string, any>> = async (data) => {
    if (!descriptor) return;
    setLoading(true);
    try {
      const res = await fetch("/api/integrations/configs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          integration_name: descriptor.name,
          display_name: descriptor.display_name,
          user_config: data,
          enabled: true,
        }),
      });

      if (!res.ok) throw new Error("Failed to create integration config");

      const newConfig = await res.json();
      onCreated(newConfig);
      onClose();
      engineWSsendMessage({
        type: "load_integration",
        data: { integration_name: descriptor.name },
      });
    } catch (err) {
      console.error("Error creating integration config:", err);
    } finally {
      setLoading(false);
    }

  };

  if (!descriptor) return null;

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add {descriptor.display_name}</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 mt-4">
          {descriptor.config_schema?.fields?.map((field) => (
            <div key={field.name} className="space-y-1">
              <Label htmlFor={field.name}>{field.label}</Label>

              {field.type === "select" ? (
                <Select
                  onValueChange={(value) => setValue(field.name, value)}
                  defaultValue={field.default}
                >
                  <SelectTrigger>
                    <SelectValue placeholder={`Select ${field.label.toLowerCase()}`} />
                  </SelectTrigger>
                  <SelectContent>
                    {field.options?.map((opt) => (
                      <SelectItem key={opt} value={opt}>
                        {opt}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              ) : field.type === "checkbox" ? (
                <div className="flex items-center space-x-2">
                  <input
                    type="checkbox"
                    id={field.name}
                    {...register(field.name)}
                    className="h-4 w-4"
                  />
                  {field.description && <span className="text-sm text-muted-foreground">{field.description}</span>}
                </div>
              ) : (
                <>
                  <Input
                    id={field.name}
                    type={field.type}
                    placeholder={field.placeholder}
                    {...register(field.name, { required: field.required })}
                  />
                  {field.description && <p className="text-sm text-muted-foreground">{field.description}</p>}
                </>
              )}
            </div>
          ))}

          <Button type="submit" className="w-full mt-4" disabled={loading}>
            {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {loading ? "Saving..." : "Create Integration"}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}
