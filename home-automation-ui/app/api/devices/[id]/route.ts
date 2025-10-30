"use server";
import { NextResponse } from "next/server";
import pool from "@/lib/db";
import { parseJsonOrFallback } from "@/lib/utils";

export async function GET(_: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  if (!id) {
    return NextResponse.json({ error: "Device ID is required" }, { status: 400 });
  }
  try {
    const [rows] = await pool.query(
      `SELECT id, integration_id, metadata, type, name, available, enabled, created_at, updated_at FROM devices WHERE id = ? LIMIT 1`,
      [id]
    );
    const row = (rows as any[])[0];
    if (!row) {
      return NextResponse.json({ error: "Device not found" }, { status: 404 });
    }
    const device = {
      id: row.id.toString(),
      integration_id: row.integration_id,
      type: row.type,
      name: row.name,
      metadata: parseJsonOrFallback(row.metadata),
      available: !!row.available,
      enabled: !!row.enabled,
      created_at: row.created_at,
      updated_at: row.updated_at,
    };
    return NextResponse.json({ device });
  } catch (error) {
    console.error("Database error (device by id):", error);
    return NextResponse.json({ error: "Failed to fetch device" }, { status: 500 });
  }
}
