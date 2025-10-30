"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import type { Device } from "@/types/device/device";
import type { Entity } from "@/types/entity/entity";
import type { State } from "@/types/state/state";
import { capitalize } from "@/lib/utils";
import { Checkbox } from "@/components/ui/checkbox";
import { ENTITY_STATE_KEY_MAP } from "@/lib/entity-display-map";
import { useSwrLite } from "@/lib/swr-lite";
import { BulkActionBar } from "@/components/common/bulk-action-bar";
import { RowActionMenu } from "@/components/common/row-action-menu";
import { EntityDetailsDialog } from "@/components/entities/entity-details-dialog";
import { SummaryStatCard } from "@/components/common/summary-stat-card";
import { Switch } from "@/components/ui/switch";

interface EntityWithState extends Entity {
  state?: State;
}

export default function DeviceDetailsPage() {
  const { id } = useParams();
  const router = useRouter();
  const [device, setDevice] = useState<Device | null>(null);
  const [loading, setLoading] = useState(true);
  const [typeFilter, setTypeFilter] = useState<string>("all");
  const [enabledTab, setEnabledTab] = useState<string>("all");
  const [selected, setSelected] = useState<string[]>([]);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailEntity, setDetailEntity] = useState<EntityWithState | null>(null);

  // SWR-lite data
  const { data: swrDevice, loading: loadingDevice } = useSwrLite<Device | null>(
    id ? `/api/devices/${id}` : null,
    async () => {
      const res = await fetch(`/api/devices/${id}`);
      if (!res.ok) throw new Error("Failed to fetch device");
      const json = await res.json();
      return (json?.device as Device) ?? null;
    },
    { ttlMs: 5 * 60_000, revalidateOnFocus: true, initialData: device ?? undefined }
  );

  const { data: swrEntities, loading: loadingEntities, mutate: mutateEntities } = useSwrLite<Entity[]>(
    id ? `/api/devices/${id}/entities` : null,
    async () => {
      const res = await fetch(`/api/devices/${id}/entities`);
      if (!res.ok) throw new Error("Failed to fetch entities");
      const json = await res.json();
      return (json?.entities as Entity[]) ?? [];
    },
    { ttlMs: 2 * 60_000, revalidateOnFocus: true }
  );

  const { data: swrStates, loading: loadingStates } = useSwrLite<State[]>(
    id ? `/api/devices/${id}/states` : null,
    async () => {
      const res = await fetch(`/api/devices/${id}/states`);
      if (!res.ok) throw new Error("Failed to fetch states");
      const json = await res.json();
      return (json?.states as State[]) ?? [];
    },
    { ttlMs: 10_000, revalidateOnFocus: true, revalidateIntervalMs: 10_000 }
  );

  // Merge entities with states first so it's available to effects below
  const entities: EntityWithState[] = useMemo(() => {
    const ents = swrEntities ?? [];
    const sts = swrStates ?? [];
    return ents.map(e => ({ ...e, state: sts.find(s => s.entity_id === e.entity_id) }));
  }, [swrEntities, swrStates]);

  useEffect(() => {
    setLoading(loadingDevice || loadingEntities || loadingStates);
  }, [loadingDevice, loadingEntities, loadingStates]);

  useEffect(() => {
    if (swrDevice) setDevice(swrDevice);
  }, [swrDevice]);

  const total = entities.length;
  const online = entities.filter(e => e.available).length;
  const offline = entities.filter(e => !e.available).length;
  const enabled = entities.filter(e => e.enabled).length;
  const disabled = entities.filter(e => !e.enabled).length;

  const types = useMemo(() => {
    const set = new Set<string>(entities.map(e => e.type).filter(Boolean));
    return ["all", ...Array.from(set)];
  }, [entities]);

  useEffect(() => {
    if (!types.includes(typeFilter) && typeFilter !== "all") {
      setTypeFilter("all");
    }
  }, [types, typeFilter]);

  const filtered = useMemo(() => {
    return entities.filter(e => {
      if (typeFilter !== "all" && e.type !== typeFilter) return false;
      if (enabledTab === "enabled" && !e.enabled) return false;
      if (enabledTab === "disabled" && e.enabled) return false;
      return true;
    });
  }, [entities, typeFilter, enabledTab]);

  // Render in original order to avoid jumping when toggling enabled
  const filteredSorted = filtered;

  // Keep selection in sync with filtered
  useEffect(() => {
    setSelected((prev) => {
      const next = prev.filter((id) => filtered.some((e) => e.external_id === id));
      if (next.length !== prev.length || next.some((id, i) => id !== prev[i])) {
        return next;
      }
      return prev;
    });
  }, [filtered]);

  const allSelected = filtered.length > 0 && filtered.every(e => selected.includes(e.external_id));
  const isPartial = selected.length > 0 && !allSelected;

  const toggleAll = () => {
    if (allSelected) {
      setSelected([]);
    } else {
      setSelected(filtered.map((e) => e.external_id));
    }
  };
  const toggleOne = (id: string) => {
    setSelected((sel) => sel.includes(id) ? sel.filter((x) => x !== id) : [...sel, id]);
  };

  const onRowClick = (e: React.MouseEvent, entity: EntityWithState) => {
    // Avoid opening when clicking on interactive controls, including inside dropdown menu
    const target = e.target as HTMLElement;
    if (
      target.closest("[role=checkbox]") ||
      target.closest("button") ||
      target.closest("[data-menu]") ||
      target.closest("[role=menu]") ||
      target.hasAttribute("role") && ["menuitem","menu"].includes(target.getAttribute("role")!) ||
      target.nodeName === "INPUT"
    ) {
      return;
    }
    setDetailEntity(entity);
    setDetailOpen(true);
  };

  // Bulk actions
  async function performBulkEntityAction(action: "enable" | "disable") {
    if (selected.length === 0) return;
    const prev = entities;
    mutateEntities(() => prev.map(e => selected.includes(e.external_id) ? { ...e, enabled: action === "enable" } : e));
    setSelected([]);
    try {
      const res = await fetch("/api/entities/bulk", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action, ids: selected }),
      });
      if (!res.ok) throw new Error("Bulk entity action failed");
    } catch (e) {
      mutateEntities(() => prev); // rollback
    }
  }

  async function enableOne(entityId: string) {
    const prev = entities;
    mutateEntities(() => prev.map(e => e.external_id === entityId ? { ...e, enabled: true } : e));
    try {
      const res = await fetch("/api/entities/bulk", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "enable", ids: [entityId] }),
      });
      if (!res.ok) throw new Error("Failed to enable entity");
    } catch (e) {
      mutateEntities(() => prev);
    }
  }
  async function disableOne(entityId: string) {
    const prev = entities;
    mutateEntities(() => prev.map(e => e.external_id === entityId ? { ...e, enabled: false } : e));
    try {
      const res = await fetch("/api/entities/bulk", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "disable", ids: [entityId] }),
      });
      if (!res.ok) throw new Error("Failed to disable entity");
    } catch (e) {
      mutateEntities(() => prev);
    }
  }

  if (loading || !device) {
    return (
      <div className="container py-6 space-y-6">
        <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="p-4 h-20 rounded-md w-full" />
          ))}
        </div>
        <div className="flex flex-col gap-3 pt-2">
          <div className="flex flex-wrap items-center gap-2">
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
                <TableHead>Name</TableHead>
                <TableHead>Entity ID</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Last Updated</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {Array.from({ length: 6 }).map((_, i) => (
                <TableRow key={i}>
                  <TableCell><Skeleton className="w-32 h-6 rounded" /></TableCell>
                  <TableCell><Skeleton className="w-40 h-6 rounded" /></TableCell>
                  <TableCell><Skeleton className="w-20 h-6 rounded" /></TableCell>
                  <TableCell><Skeleton className="w-24 h-6 rounded" /></TableCell>
                  <TableCell><Skeleton className="w-28 h-6 rounded" /></TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </div>
    );
  }

  if (!device) {
    return (
      <div className="container py-6">
        <Card>
          <CardHeader>
            <CardTitle>Device not found</CardTitle>
          </CardHeader>
        </Card>
      </div>
    );
  }

  return (
    <div className="container py-6 space-y-8">
      {/* Header */}
      <div className="flex flex-col gap-2">
        <h1 className="text-3xl font-semibold">{device.name}</h1>
        <div className="flex items-center gap-2">
          <Badge variant="secondary">{device.type}</Badge>
          <Badge variant={device.available ? "default" : "destructive"}>
            {device.available ? "Online" : "Offline"}
          </Badge>
        </div>
      </div>

      {/* Summary */}
      <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
        <SummaryStatCard label="Entities" value={total} />
        <SummaryStatCard label="Enabled" value={enabled} />
        <SummaryStatCard label="Disabled" value={disabled} />
      </div>

      <h2 className="text-2xl font-semibold mt-8 mb-3">Entities</h2>

      {/* Filters */}
      <div className="flex flex-col gap-3 pt-2">
        <div className="flex flex-wrap items-center gap-2">
          <Select value={typeFilter} onValueChange={setTypeFilter}>
            <SelectTrigger aria-label="Entity type">
              <SelectValue placeholder="Select type" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectLabel>Entity Types</SelectLabel>
                {types.map((t) => (
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
          <div className="ml-auto flex items-center gap-2">
            <BulkActionBar selectedCount={selected.length} onEnable={() => performBulkEntityAction("enable")} onDisable={() => performBulkEntityAction("disable")} />
          </div>
        </div>
      </div>

      {/* Entities table */}
      <div className="bg-card/50 rounded-lg overflow-x-auto mt-2">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-10 p-0">
                <Checkbox
                  checked={isPartial ? 'indeterminate' : allSelected}
                  onCheckedChange={toggleAll}
                  aria-label="Select all entities"
                  className="mx-auto"
                />
              </TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Entity ID</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>State</TableHead>
              <TableHead>Enabled</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filtered.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center py-8">
                  <span className="text-muted-foreground">No entities match your filter.</span>
                </TableCell>
              </TableRow>
            ) : (
              filtered.map(entity => (
                <TableRow key={entity.external_id} onClick={(e) => onRowClick(e, entity)} className="cursor-pointer">
                  <TableCell className="p-0">
                    <Checkbox
                      checked={selected.includes(entity.external_id)}
                      onCheckedChange={() => toggleOne(entity.external_id)}
                      aria-label={`Select entity ${entity.name}`}
                      className="mx-auto"
                    />
                  </TableCell>
                  <TableCell className="font-medium">{entity.name}</TableCell>
                  <TableCell>{entity.entity_id}</TableCell>
                  <TableCell>{capitalize(entity.type)}</TableCell>
                  <TableCell>
                    {(() => {
                      const s = entity.state?.state;
                      const map = ENTITY_STATE_KEY_MAP[entity.type];
                      if (typeof s === "boolean") {
                        const off = map?.falseLabel || "FALSE";
                        const on = map?.trueLabel || "TRUE";
                        return map?.label
                          ? `${map.label}: ${s ? on : off}`
                          : s ? on : off;
                      }
                      if (typeof s === "string" || typeof s === "number") {
                        return (map?.label ? `${map.label}: ` : "") + String(s);
                      }
                      return <span className="text-muted-foreground">–</span>;
                    })()}
                  </TableCell>
                  <TableCell>
                    <Switch
                      checked={entity.enabled}
                      onCheckedChange={(checked) => {
                        if (checked) {
                          enableOne(entity.external_id);
                        } else {
                          disableOne(entity.external_id);
                        }
                      }}
                      onClick={(e) => e.stopPropagation()}
                      aria-label={`Toggle ${entity.name}`}
                    />
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Details Dialog */}
      <EntityDetailsDialog open={detailOpen} onOpenChange={setDetailOpen} entity={detailEntity} />
    </div>
  );
}
