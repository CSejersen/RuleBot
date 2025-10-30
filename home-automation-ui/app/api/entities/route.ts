"use server";
// app/api/entities/route.ts
import { NextResponse } from "next/server"
import pool from "@/lib/db"
import { Entity } from "@/types/entity"

export async function GET() {
  try {
    const [rows] = await pool.query(`
      SELECT external_id, device_id, entity_id, type, name, available, enabled
      FROM entities
      ORDER BY name ASC
    `)

    const entities: Entity[] = (rows as any[]).map(row => ({
      external_id: row.id,
      device_id: row.device_id,
      entity_id: row.entity_id,
      type: row.type,
      name: row.name,
      available: !!row.available,
      enabled: !!row.enabled,
    }))

    return NextResponse.json({ entities })
  } catch (error) {
    console.error("Database error:", error)
    return NextResponse.json({ error: "Failed to fetch entities" }, { status: 500 })
  }
}
