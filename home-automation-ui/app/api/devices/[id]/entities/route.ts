"use server";
import { NextResponse } from "next/server";
import pool from "@/lib/db";
import type { Entity } from "@/types/entity";

export async function GET(_: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const [rows] = await pool.query(`
      SELECT external_id, device_id, entity_id, type, name, available, enabled
      FROM entities
      WHERE device_id = ?
      ORDER BY name ASC
    `, [id]);

    const entities: Entity[] = (rows as any[]).map(row => ({
      external_id: row.external_id,
      device_id: row.device_id,
      entity_id: row.entity_id,
      type: row.type,
      name: row.name,
      available: !!row.available,
      enabled: !!row.enabled,
    }));

    return NextResponse.json({ entities: entities });
  } catch (error) {
    console.error(`Database error for device ${id}:`, error);
    return NextResponse.json({ error: "Failed to fetch entities" }, { status: 500 });
  }
}
