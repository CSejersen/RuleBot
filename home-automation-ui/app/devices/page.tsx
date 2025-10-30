"use client";

import { useEffect, useMemo, useState } from "react";
import { Loader2, Plus, Square } from "lucide-react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableHeader, TableBody, TableHead, TableRow, TableCell } from "@/components/ui/table";
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import type { Device } from "@/types/device";
import type { IntegrationConfig } from "@/types/integration/integration-config";
import { capitalize } from "@/lib/utils";
import { Checkbox } from "@/components/ui/checkbox";
import { useSwrLite } from "@/lib/swr-lite";
import { BulkActionBar } from "@/components/common/bulk-action-bar";
import { RowActionMenu } from "@/components/common/row-action-menu";
import { SummaryStatCard } from "@/components/common/summary-stat-card";

export default function DevicesPage() {
  const [integrationFilter, setIntegrationFilter] = useState<string>("all");
  const [typeTab, setTypeTab] = useState<string>("all");
  const [enabledTab, setEnabledTab] = useState<string>("all");
  const [selected, setSelected] = useState<string[]>([]);
  const router = useRouter();

  const { data: devicesData, loading: loadingDevices, mutate: mutateDevices } = useSwrLite<Device[]>(
    "/api/devices",
    async () => {
      const res = await fetch("/api/devices");
      if (!res.ok) throw new Error("Failed to fetch devices");
      const json = (await res.json()) as { devices: Device[] };
      return json.devices ?? [];
    },
    { ttlMs: 5 * 60_000, revalidateOnFocus: true }
  );

  const { data: configsData, loading: loadingConfigs } = useSwrLite<IntegrationConfig[]>(
    "/api/integrations/configs",
    async () => {
      const res = await fetch("/api/integrations/configs");
      if (!res.ok) throw new Error("Failed to fetch integration configs");
      return (await res.json()) as IntegrationConfig[];
    },
    { ttlMs: 10 * 60_000, revalidateOnFocus: true }
  );

  const loading = loadingDevices || loadingConfigs;
  const devices = devicesData ?? [];
  const integrationConfigs = configsData ?? [];

  // Smart summary stats
  const total = devices.length;
  const online = devices.filter((d) => d.available).length;
  const offline = devices.filter((d) => !d.available).length;
  const disabled = devices.filter((d) => !d.enabled).length;

  const integrationOptions = useMemo(() => {
    return [{ id: "all", name: "All integrations" }, ...integrationConfigs.map(c => ({ id: String(c.id), name: c.display_name }))];
  }, [integrationConfigs]);

  const deviceTypes = useMemo(() => {
    const set = new Set<string>(
      devices
        .filter(d => integrationFilter === "all" || String(d.integration_id) === integrationFilter)
        .map(d => d.type)
        .filter(Boolean)
    );
    return ["all", ...Array.from(set)];
  }, [devices, integrationFilter]);

  // Ensure current type selection is valid when integration changes
  useEffect(() => {
    if (!deviceTypes.includes(typeTab) && typeTab !== "all") {
      setTypeTab("all");
    }
  }, [deviceTypes, typeTab]);

  const filtered = devices.filter((d) => {
    if (integrationFilter !== "all" && String(d.integration_id) !== integrationFilter) return false;
    if (typeTab !== "all" && d.type !== typeTab) return false;
    if (enabledTab === "enabled" && !d.enabled) return false;
    if (enabledTab === "disabled" && d.enabled) return false;
    return true;
  });

  const filteredSorted = useMemo(
    () => [...filtered].sort((a, b) => (a.enabled === b.enabled ? 0 : a.enabled ? -1 : 1)),
    [filtered]
  );

  // Update selection if list changes (deselect non-visible devices)
  useEffect(() => {
    setSelected((prev) => {
      const next = prev.filter((id) => filtered.some((d) => d.id === id));
      if (next.length !== prev.length || next.some((id, i) => id !== prev[i])) {
        return next;
      }
      return prev;
    });
  }, [filtered]);

  const allSelected = filtered.length > 0 && filtered.every((d) => selected.includes(d.id));
  const isPartial = selected.length > 0 && !allSelected;
  const toggleAll = () => {
    if (allSelected) {
      setSelected([]);
    } else {
      setSelected(filtered.map((d) => d.id));
    }
  };
  const toggleOne = (id: string) => {
    setSelected((sel) => sel.includes(id) ? sel.filter((x) => x !== id) : [...sel, id]);
  };

  // Bulk actions
  const enableSelected = async () => {
    if (selected.length === 0) return;
    const prev = devices;
    mutateDevices((p?: Device[]) => (p ?? []).map((d: Device) => (selected.includes(d.id) ? { ...d, enabled: true } : d)));
    setSelected([]);
    try {
      const res = await fetch("/api/devices/bulk", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "enable", ids: selected }),
      });
      if (!res.ok) throw new Error("Bulk enable failed");
    } catch (e) {
      mutateDevices(() => prev);
    }
  };
  const disableSelected = async () => {
    if (selected.length === 0) return;
    const prev = devices;
    mutateDevices((p?: Device[]) => (p ?? []).map((d: Device) => (selected.includes(d.id) ? { ...d, enabled: false } : d)));
    setSelected([]);
    try {
      const res = await fetch("/api/devices/bulk", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "disable", ids: selected }),
      });
      if (!res.ok) throw new Error("Bulk disable failed");
    } catch (e) {
      mutateDevices(() => prev);
    }
  };
  const deleteSelected = async () => {
    if (selected.length === 0) return;
    const ok = window.confirm(`Delete ${selected.length} selected device(s)? This cannot be undone.`);
    if (!ok) return;
    const prev = devices;
    mutateDevices((p?: Device[]) => (p ?? []).filter((d: Device) => !selected.includes(d.id)));
    const ids = selected;
    setSelected([]);
    try {
      const res = await fetch("/api/devices/bulk", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "delete", ids }),
      });
      if (!res.ok) throw new Error("Bulk delete failed");
    } catch (e) {
      mutateDevices(() => prev);
    }
  };

  // Helper to perform bulk actions, reused for single-row actions
  async function performAction(action: "enable" | "disable" | "delete", ids: string[], optimistic: () => void, rollback: () => void) {
    optimistic();
    try {
      const res = await fetch("/api/devices/bulk", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action, ids }),
      });
      if (!res.ok) throw new Error(`Action ${action} failed`);
    } catch (e) {
      rollback();
    }
  }

  const enableOne = (id: string) => {
    const prev = devices;
    performAction(
      "enable",
      [id],
      () => mutateDevices((p?: Device[]) => (p ?? []).map((d: Device) => (d.id === id ? { ...d, enabled: true } : d))),
      () => mutateDevices(() => prev)
    );
  };

  const disableOne = (id: string) => {
    const prev = devices;
    performAction(
      "disable",
      [id],
      () => mutateDevices((p?: Device[]) => (p ?? []).map((d: Device) => (d.id === id ? { ...d, enabled: false } : d))),
      () => mutateDevices(() => prev)
    );
  };

  const deleteOne = (id: string, name: string) => {
    const ok = window.confirm(`Delete device \"${name}\"? This cannot be undone.`);
    if (!ok) return;
    const prev = devices;
    performAction(
      "delete",
      [id],
      () => mutateDevices((p?: Device[]) => (p ?? []).filter((d: Device) => d.id !== id)),
      () => mutateDevices(() => prev)
    );
  };

  // Device icon, fallback to generic square for now
  const typeIcon = (t: string) => <Square className="w-5 h-5 text-muted-foreground mr-1" />;

  return (
    <div className="container py-6 space-y-8">
      {/* HEADER & SUMMARY */}
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between pb-4">
        <h1 className="text-3xl font-semibold">Devices</h1>
        <Button onClick={() => router.push("/devices/new")} className="gap-2">
          <Plus className="w-4 h-4" /> Add Device
        </Button>
      </div>
      {loading ? (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="p-4 h-20 rounded-md w-full" />
            ))}
          </div>
          <div className="flex flex-col gap-3 pt-2">
            <div className="flex flex-wrap items-center gap-2">
              <Skeleton className="h-9 w-64 rounded-md" />
              <Skeleton className="h-9 w-48 rounded-md" />
            </div>
            <div className="flex flex-wrap items-center gap-4">
              <Skeleton className="h-9 w-40 rounded-md" />
            </div>
          </div>
          <div className="bg-card/50 rounded-lg overflow-x-auto mt-2">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead></TableHead>
                  <TableHead>Name</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Integration</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {Array.from({ length: 6 }).map((_, i) => (
                  <TableRow key={i}>
                    <TableCell><Skeleton className="w-6 h-6 rounded" /></TableCell>
                    <TableCell><Skeleton className="w-32 h-6 rounded" /></TableCell>
                    <TableCell><Skeleton className="w-20 h-6 rounded" /></TableCell>
                    <TableCell><Skeleton className="w-24 h-6 rounded" /></TableCell>
                    <TableCell><Skeleton className="w-20 h-6 rounded" /></TableCell>
                    <TableCell><Skeleton className="w-10 h-6 rounded" /></TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <SummaryStatCard label="Total Devices" value={total} />
            <SummaryStatCard label="Online" value={online} />
            <SummaryStatCard label="Offline" value={offline} />
            <SummaryStatCard label="Disabled" value={disabled} />
          </div>

          {/* FILTERS: Integration Select + Type Select + Enabled Tabs */}
          <div className="flex flex-col gap-3 pt-2">
            <div className="flex flex-wrap items-center gap-2">
              <Select value={integrationFilter} onValueChange={setIntegrationFilter}>
                <SelectTrigger aria-label="Integration">
                  <SelectValue placeholder="Select integration" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectLabel>Integrations</SelectLabel>
                    {integrationOptions.map((opt) => (
                      <SelectItem key={opt.id} value={opt.id}>{opt.name}</SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>

              <Select value={typeTab} onValueChange={setTypeTab}>
                <SelectTrigger aria-label="Device type">
                  <SelectValue placeholder="Select type" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectLabel>Device Types</SelectLabel>
                    {deviceTypes.map((t) => (
                      <SelectItem key={t} value={t}>{t === "all" ? "All Types" : `${capitalize(t)}s`}</SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-wrap items-center gap-4">
              <Tabs value={enabledTab} onValueChange={setEnabledTab}>
                <TabsList>
                  <TabsTrigger value="all">All</TabsTrigger>
                  <TabsTrigger value="enabled">Enabled</TabsTrigger>
                  <TabsTrigger value="disabled">Disabled</TabsTrigger>
                </TabsList>
              </Tabs>
              <BulkActionBar selectedCount={selected.length} onEnable={enableSelected} onDisable={disableSelected} onDelete={deleteSelected} />
            </div>
          </div>

          {/* TABLE VIEW */}
          <div className="bg-card/50 rounded-lg overflow-x-auto mt-2">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-10 p-0">
                    <Checkbox
                      checked={isPartial ? 'indeterminate' : allSelected}
                      onCheckedChange={toggleAll}
                      aria-label="Select all"
                      className="mx-auto"
                    />
                  </TableHead>
                  <TableHead>Name</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Integration</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredSorted.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-8">
                      <span className="text-muted-foreground">No devices match your filter.</span>
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredSorted.map((device) => {
                    const config = integrationConfigs.find((c) => c.id === device.integration_id);
                    return (
                      <TableRow key={device.id}>
                        <TableCell className="p-0">
                          <Checkbox
                            checked={selected.includes(device.id)}
                            onCheckedChange={() => toggleOne(device.id)}
                            aria-label={`Select device ${device.name}`}
                            className="mx-auto"
                          />
                        </TableCell>
                        <TableCell className="font-medium cursor-pointer hover:underline" onClick={() => router.push(`/devices/${device.id}`)}>{device.name}</TableCell>
                        <TableCell>{capitalize(device.type)}</TableCell>
                        <TableCell>{config?.display_name ?? device.integration_id}</TableCell>
                        <TableCell>
                          <Badge variant={device.available ? 'default' : 'destructive'}>{device.available ? 'Online' : 'Offline'}</Badge>
                          {!device.enabled && (
                            <Badge variant="secondary" className="ml-2">Disabled</Badge>
                          )}
                        </TableCell>
                        <TableCell>
                          <RowActionMenu onEnable={() => enableOne(device.id)} onDisable={() => disableOne(device.id)} onDetails={() => router.push(`/devices/${device.id}`)} onDelete={() => deleteOne(device.id, device.name)} />
                        </TableCell>
                      </TableRow>
                    );
                  })
                )}
              </TableBody>
            </Table>
          </div>
        </>
      )}
    </div>
  );
}
