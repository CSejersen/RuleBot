"use client";

import { useEffect, useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Loader2, Plug, Radar, RefreshCw } from "lucide-react";
import { IntegrationDescriptor } from "@/types/integration/integration-descriptor"
import { IntegrationConfig } from "@/types/integration/integration-config";
import { AddIntegrationDialog } from "./add-integration-dialog";
import { DiscoveryButton } from "./discovery_button";


export default function IntegrationsPage() {
  const [descriptors, setDescriptors] = useState<IntegrationDescriptor[]>([]);
  const [configs, setConfigs] = useState<IntegrationConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedIntegration, setSelectedIntegration] = useState<IntegrationDescriptor | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);

  useEffect(() => {
    async function fetchData() {
      try {
        const [descRes, cfgRes] = await Promise.all([
          fetch("/api/integrations/descriptors"),
          fetch("/api/integrations/configs"),
        ]);
        setDescriptors(await descRes.json());
        setConfigs(await cfgRes.json());
      } catch (e) {
        console.error("Failed to load integration", e);
      } finally {
        setLoading(false);
      }
    }
    fetchData();
  }, []);

  if (loading) {
    return (
      <div className="flex justify-center items-center h-full">
        <Loader2 className="animate-spin" />
      </div>
    );
  }

  const enrichedConfigs = configs.map((cfg) => {
    const desc = descriptors.find((d) => d.name === cfg.integration_name);
    return {
      ...cfg,
      descriptor: desc || null,
    };
  });

  return (
    <div className="container py-6">
      <h1 className="text-3xl font-semibold mb-6">Integrations</h1>

      <Tabs defaultValue="configured">
        <TabsList>
          <TabsTrigger value="configured">Configured</TabsTrigger>
          <TabsTrigger value="available">Available</TabsTrigger>
        </TabsList>

        {/* Configured Integrations */}
        <TabsContent value="configured" className="grid md:grid-cols-2 lg:grid-cols-3 gap-4 mt-4">
          {enrichedConfigs.length === 0 && (
            <p className="text-muted-foreground mt-4">No integrations configured yet.</p>
          )}
          {enrichedConfigs.map((cfg) => (
            <Card key={cfg.id}>
              <CardHeader>
                <CardTitle>{cfg.display_name}</CardTitle>
                <CardDescription>
                  {cfg.descriptor?.description && (
                    <p className="text-sm text-muted-foreground mt-1">
                      {cfg.descriptor.description}
                    </p>
                  )}
                </CardDescription>
              </CardHeader>

              {cfg.descriptor && (
                <CardContent className="flex flex-wrap gap-2">
                  {cfg.descriptor.capabilities.map((cap) => (
                    <Badge key={cap}>{cap}</Badge>
                  ))}
                </CardContent>
              )}

              <CardFooter className="flex justify-between">
                <Button size="sm" variant="outline">
                  <Plug className="w-4 h-4 mr-1" /> Edit
                </Button>
                {cfg.descriptor?.capabilities.includes("discovery") && (
                  <DiscoveryButton integrationName={cfg.integration_name} />
                )}
              </CardFooter>
            </Card>
          ))}
        </TabsContent>

        {/* Available Integrations */}
        <TabsContent
          value="available"
          className="grid md:grid-cols-2 lg:grid-cols-3 gap-4 mt-4"
        >
          {descriptors.map((desc) => {
            const isConfigured = configs.some(
              (cfg) => cfg.integration_name === desc.name
            );

            return (
              <Card key={desc.name}>
                <CardHeader>
                  <CardTitle>{desc.display_name}</CardTitle>
                  <CardDescription>{desc.description}</CardDescription>
                </CardHeader>

                <CardContent className="flex flex-wrap gap-2">
                  {desc.capabilities.map((cap) => (
                    <Badge variant="outline" key={cap}>
                      {cap}
                    </Badge>
                  ))}
                </CardContent>

                <CardFooter>
                  <Button
                    onClick={() => {
                      if (!isConfigured) {
                        setSelectedIntegration(desc);
                        setDialogOpen(true);
                      }
                    }}
                    disabled={isConfigured}
                    variant={isConfigured ? "secondary" : "default"}
                    className="w-full"
                  >
                    {isConfigured ? (
                      <>
                        <Plug className="w-4 h-4 mr-2 opacity-60" />
                        Already Configured
                      </>
                    ) : (
                      <>
                        <Plug className="w-4 h-4 mr-2" />
                        Add Integration
                      </>
                    )}
                  </Button>
                </CardFooter>
              </Card>
            );
          })}
        </TabsContent>
      </Tabs>

      <AddIntegrationDialog
        open={dialogOpen}
        onClose={() => setDialogOpen(false)}
        descriptor={selectedIntegration}
        onCreated={(newConfig: IntegrationConfig) => {
          setConfigs((prev) => [newConfig, ...prev]);
          setDialogOpen(false);
        }}
      />
    </div>
  );
}
