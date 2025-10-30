"use server";
import { NextResponse } from "next/server";
import pool from "@/lib/db";
import { parseJsonOrFallback } from "@/lib/utils";

export async function GET() {
  try {
    const [rows] = await pool.query(`
      SELECT id, integration_id, metadata, type, name, available, enabled, created_at, updated_at
      FROM devices
      ORDER BY name ASC
    `)
    // Shape output for JSON, but don't cast as Device[]
    const devices = (rows as any[]).map(row => ({
      id: row.id.toString(),
      integration_id: row.integration_id,
      type: row.type,
      name: row.name,
      metadata: parseJsonOrFallback(row.metadata),
      available: !!row.available,
      enabled: !!row.enabled,
      created_at: row.created_at,
      updated_at: row.updated_at,
    }))
    return NextResponse.json({ devices })
  } catch (error) {
    console.error("Database error:", error)
    return NextResponse.json({ error: "Failed to fetch entities" }, { status: 500 })
  }
}

