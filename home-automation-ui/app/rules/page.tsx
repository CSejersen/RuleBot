"use client"

import { useEffect, useState } from "react"
import { columns } from "./data-table/columns"
import { Rule } from "@/app/api/rules/types/rule"
import { DataTable } from "./data-table/data-table"
import { Button } from "@/components/ui/button"
import { CreateRuleSheet } from "./create-rule-sheet"

async function getData(): Promise<Rule[]> {
  const res = await fetch("/api/rules", { cache: "no-store" })
  if (!res.ok) throw new Error("Failed to fetch rules")
  return res.json()
}

export default function RulesPage() {
  const [rules, setRules] = useState<Rule[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isSheetOpen, setIsSheetOpen] = useState(false)

  useEffect(() => {
    getData()
      .then(setRules)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div>Loading rules...</div>
  if (error) return <div>Error: {error}</div>

  return (
    <div className="container mx-auto py-10">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-semibold tracking-tight">Rules</h2>
        <Button size="sm" onClick={() => setIsSheetOpen(true)}>
          Create New Rule
        </Button>
      </div>

      <DataTable columns={columns} data={rules} />

      <CreateRuleSheet
        open={isSheetOpen}
        onOpenChange={setIsSheetOpen}
        onRuleCreated={(rule) => setRules((prev) => [...prev, rule])}
      />
    </div>
  )
}
