"use client"

import { useEffect, useState, useCallback } from "react"
import { columns } from "./data-table/columns"
import { Automation } from "@/types/automation/automation"
import { DataTable } from "./data-table/data-table"
import { Button } from "@/components/ui/button"
import { CreateAutomationSheet } from "./create-automation-sheet"
import { AutomationDetailsDialog } from "./data-table/automation-details-dialog"
import { RowActionMenu } from "@/components/common/row-action-menu"
import { useMemo } from "react"
import { engineWSsendMessage } from "@/lib/engine-socket"

async function getData(): Promise<Automation[]> {
  const res = await fetch("/api/automations", { cache: "no-store" })
  if (!res.ok) throw new Error("Failed to fetch automations")
  return res.json()
}

export default function AutomationsPage() {
  if (typeof window !== "undefined") console.count("Render: AutomationsPage")
  const [automations, setAutomations] = useState<Automation[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isSheetOpen, setIsSheetOpen] = useState(false)

  // Viewing/editing automation details
  const [selectedAutomation, setSelectedAutomation] = useState<Automation | null>(null)
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [editingAutomation, setEditingAutomation] = useState<Automation | null>(null)

  useEffect(() => {
    getData()
      .then(setAutomations)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [])

  const columnsWithActions = useMemo(() => {
    return [
      ...columns,
      {
        id: "actions",
        header: "",
        cell: ({ row }: any) => (
          <div className="flex justify-end">
            <RowActionMenu
              onDetails={() => {
                setSelectedAutomation(row.original)
                setIsDialogOpen(true)
              }}
              onEdit={() => {
                setEditingAutomation(row.original)
                setIsDialogOpen(false)
                setIsSheetOpen(true)
              }}
              onDelete={async () => {
                const automation = row.original as Automation
                if (!automation?.id) return
                await deleteAutomationOptimistic(automation)
              }}
            />
          </div>
        ),
      } as any,
    ]
  }, [columns])

  const handleRowClick = useCallback((automation: Automation) => {
    setSelectedAutomation(automation)
    setIsDialogOpen(true)
  }, [])

  const deleteAutomationOptimistic = useCallback(async (automation: Automation) => {
    if (typeof window !== "undefined") console.log("deleteAutomationOptimistic called", automation.id)
    let snapshot: Automation[] = []
    // optimistic remove with snapshot capture
    setAutomations((prev) => {
      if (typeof window !== "undefined") console.log("optimistic remove setAutomations", { prevLength: prev.length })
      snapshot = prev
      return prev.filter((a) => a.id !== automation.id)
    })
    try {
      const res = await fetch(`/api/automations/${automation.id}`, { method: "DELETE" })
      if (!res.ok) {
        const data = await res.json().catch(() => ({}))
        throw new Error(data.error || "Failed to delete automation")
      }
      if (typeof window !== "undefined") console.log("delete success", automation.id)
      engineWSsendMessage({ type: "reload_automations" })
    } catch (err: any) {
      // rollback
      if (typeof window !== "undefined") console.log("delete failed, rollback", automation.id)
      setAutomations(snapshot)
      alert(err?.message || "An error occurred while deleting the automation")
    }
  }, [])


  if (loading) return <div>Loading automations...</div>
  if (error) return <div>Error: {error}</div>

  return (
    <div className="container mx-auto py-10">
      {/* Header */}
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-semibold tracking-tight">Automations</h2>
        <Button
          size="sm"
          onClick={() => {
            setEditingAutomation(null) // Creating a new automation
            setIsSheetOpen(true)
          }}
        >
          Create New Automation
        </Button>
      </div>

      {/* Automations Table */}
      <DataTable
        columns={columnsWithActions as any}
        data={automations}
        onRowClick={handleRowClick}
      />

      {/* Create/Edit Automation Sheet */}
      <CreateAutomationSheet
        key={editingAutomation?.id ?? "new"}
        open={isSheetOpen}
        onOpenChange={(open) => {
          setIsSheetOpen(open)
          if (open) {
            // Ensure details dialog does not fight the sheet
            setIsDialogOpen(false)
            setSelectedAutomation(null)
          }
        }}
        editingAutomation={editingAutomation}
        onAutomationSaved={(automation) => {
          // Normalize shape to prevent undefined rows/fields
          const normalized: Automation = {
            id: (automation as any).id,
            alias: automation.alias ?? "",
            description: automation.description ?? "",
            triggers: automation.triggers ?? [],
            conditions: (automation as any).conditions ?? [],
            actions: automation.actions ?? [],
            enabled: automation.enabled ?? true,
            last_triggered: (automation as any).last_triggered ?? null,
          } as Automation
          setAutomations((prev) => {
            const index = prev.findIndex((a) => a.id === normalized.id)
            if (index !== -1) {
              const updated = [...prev]
              updated[index] = normalized
              return updated
            }
            return [...prev, normalized]
          })
          setEditingAutomation(null)
        }}
      />

      {/* Automation Details Dialog */}
      <AutomationDetailsDialog
        automation={selectedAutomation}
        open={isDialogOpen}
        onOpenChange={(open) => {
          setIsDialogOpen(open)
          if (!open) {
            setSelectedAutomation(null)
          }
        }}
        onEdit={(automation) => {
          setEditingAutomation(automation) // Pass automation to sheet for editing
          setIsDialogOpen(false)
          setIsSheetOpen(true)
        }}
        onDeleteOptimistic={deleteAutomationOptimistic}
      />
    </div>
  )
}
